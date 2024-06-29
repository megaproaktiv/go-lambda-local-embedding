---
title: "Getting a near-real-time view of a DynamoDB stream with Python"
author: "Maurice Borgmeier"
date: 2022-05-27
toc: false
draft: false
image: "img/2022/05/towfiqu-barbhuiya-xkArbdUcUeE-unsplash.jpg"
thumbnail: "img/2022/05/towfiqu-barbhuiya-xkArbdUcUeE-unsplash.jpg"
categories: ["aws"]
tags: ["level-400", "dynamodb", "python"]

---

DynamoDB streams help you respond to changes in your tables, which is commonly used to create aggregations or trigger other workflows once data is updated. Getting a near-real-time view into these Streams can also be helpful during developing or debugging a Serverless application in AWS. Today, I will share a [Python script](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-streamgazer) that I built to hook into DynamoDB streams.

Before we begin, I suggest you read my blog post that [contains a deep dive into DynamoDB streams and how they're implemented](https://dev.to/aws-builders/deep-dive-into-dynamodb-streams-and-the-lambda-integration-gjn) because we'll be using these concepts today. To summarize, DynamoDB tables consist of storage partitions to which shards attach, which make up the stream. We can read records from these shards and process them any way we like.

Our goal is to create a tool that can do precisely that and display changes in near-real-time. To build our client in Python, we need to begin by listing all the shards in the stream, which requires us to recursively call the `DescribeStream` API as boto3 doesn't have a paginator for this operation (yet).

```python
Shard = collections.namedtuple(
    typename="Shard",
    field_names=[
        "stream_arn",
        "shard_id",
        "parent_shard_id",
        "starting_sequence_number",
        "ending_sequence_number"
    ]
)

def list_all_shards(stream_arn: str, **kwargs: dict) -> typing.List[Shard]:

    def _shard_response_to_shard(response: dict) -> Shard:
        return Shard(
            stream_arn=stream_arn,
            shard_id=response.get("ShardId"),
            parent_shard_id=response.get("ParentShardId"),
            starting_sequence_number=response.get("SequenceNumberRange", {}).get("StartingSequenceNumber"),
            ending_sequence_number=response.get("SequenceNumberRange", {}).get("EndingSequenceNumber")
        )
 
    client = boto3.client("dynamodbstreams")
    pagination_args = {}
    exclusive_start_shard_id = kwargs.get("next_page_identifier", None)
    if exclusive_start_shard_id is not None:
        pagination_args["ExclusiveStartShardId"] = exclusive_start_shard_id
    
    response = client.describe_stream(
        StreamArn=stream_arn,
        **pagination_args
    )

    list_of_shards = [_shard_response_to_shard(item) for item in response["StreamDescription"]["Shards"]]

    next_page_identifier = response["StreamDescription"].get("LastEvaluatedShardId")
    if next_page_identifier is not None:
        list_of_shards += list_all_shards(
            stream_arn=stream_arn,
            next_page_identifier=next_page_identifier
        )
    
    return list_of_shards
```

I chose to create a little class called `Shard` to encapsulate the concept of a shard using a `namedtuple` from the [collections](https://docs.python.org/3/library/collections.html#collections.namedtuple) module. Now that we have a list of shards, we only care about those not yet closed because we want a near-real-time view of **current** events. Closed shards have an `EndingSequenceNumber` so that we can filter them out like this.

```python
def is_open_shard(shard: Shard) -> bool:
    return shard.ending_sequence_number is None

def list_open_shards(stream_arn: str) -> typing.List[Shard]:
    all_shards = list_all_shards(
        stream_arn=stream_arn
    )

    open_shards = [shard for shard in all_shards if is_open_shard(shard)]

    return open_shards
```

We want to request all the records in each of these shards, which we do by creating a shard iterator and then using that to retrieve records. The `GetRecords` API also returns a new shard iterator that we can use for our subsequent request. If there is no new shard iterator in the response, it means that the shard is closed.

```python
def get_shard_iterator(shard: Shard, iterator_type: str = "LATEST") -> str:
    client = boto3.client("dynamodbstreams")

    response = client.get_shard_iterator(
        StreamArn=shard.stream_arn,
        ShardId=shard.shard_id,
        ShardIteratorType=iterator_type
    )
    
    return response["ShardIterator"]

def get_next_records(shard_iterator: str) -> typing.Tuple[typing.List[dict], str]:
    client = boto3.client("dynamodbstreams")

    response = client.get_records(
        ShardIterator=shard_iterator
    )

    return response["Records"], response.get("NextShardIterator")
```

Putting this together means creating a `shard_watcher` function to fetch the most recent records from a particular shard periodically. This function receives the shard it's responsible for and a list of functions that will be called with each record it receives. You can think of them as Observers and the records being the Observable if you're familiar with the [Observer pattern](https://en.wikipedia.org/wiki/Observer_pattern). The optional parameter `start_at_oldest` controls whether the shard will be watched from the oldest available record or the most recent one. We also wait a little bit in the loop before requesting new records. This is to avoid hammering the AWS API too much.

```python
def shard_watcher(shard: Shard, callables: typing.List[typing.Callable], start_at_oldest = False):
    
    shard_iterator_type = "TRIM_HORIZON" if start_at_oldest else "LATEST"
    shard_iterator = get_shard_iterator(shard, shard_iterator_type)

    while shard_iterator is not None:
        records, shard_iterator = get_next_records(shard_iterator)

        for record in records:
            for handler in callables:
                handler(record)
        
        time.sleep(0.5)
```

This allows us to watch a single shard, but in reality, a stream comprises multiple shards, and we need to watch all of them, so we don't miss changes. That's why I implemented a function to manage the watchers. It receives the stream ARN and the list of observers and uses the [multiprocessing](https://docs.python.org/3/library/multiprocessing.html) module to spawn a watcher process for each shard, so they're watched in parallel.

```python
def start_watching(stream_arn: str, callables: typing.List[typing.Callable]) -> None:

    shard_to_watcher: typing.Dict[str, mp.Process] = {}
    initial_loop = True

    while True:

        open_shards = list_open_shards(stream_arn=stream_arn)
        start_at_oldest = True
        if initial_loop:
            start_at_oldest = False
            initial_loop = False

        for shard in open_shards:
            if shard.shard_id not in shard_to_watcher:

                print("Starting watcher for shard:", shard.shard_id)
                args = (shard, callables, start_at_oldest)
                process = mp.Process(target=shard_watcher, args=args)
                shard_to_watcher[shard.shard_id] = process
                process.start()
                
        time.sleep(10)
```

This function periodically lists all the shards in the stream and ensures there is a watcher for each shard. Each shard it discovers in the first loop will be followed from the most recent record when the function starts. Any newly discovered shard will be read from the oldest available record in subsequent loops. After we begin, we don't want to miss any record.

I've also implemented two basic observers that can handle change records. The first function prints a summary of the change that consists of the type of operation, the timestamp, and the item's keys. The second one is even more basic and prints the record.

```python
def print_summary(change_record: dict):

    changed_at:datetime = change_record["dynamodb"]["ApproximateCreationDateTime"]
    event_type:str = change_record["eventName"]

    item_keys:dict = change_record["dynamodb"]["Keys"]
    item_key_list = []
    for key in sorted(item_keys.keys()):
        value = item_keys[key][list(item_keys[key].keys())[0]]
        item_key_list.append(f"{key}={value}")
    
    output_str = "[{0}] - {1:^6} - {2}".format(changed_at.isoformat(timespec="seconds"), event_type, ", ".join(item_key_list))

    print(output_str)

def print_change_record(change_record: dict):
    print(change_record)
```

I've implemented an argument parser that takes the command line arguments and sets everything up accordingly to make this callable from the outside. The [argparse](https://docs.python.org/3/library/argparse.html) module from the standard library is instrumental here.

```python
def main():

    parser = argparse.ArgumentParser(description="See what's going on in DynamoDB Streams in near real-time 🔍")
    parser.add_argument("stream_arn", type=str, help="The ARN of the stream you want to watch.")
    parser.add_argument("--print-record", "-pr", action="store_true", help="Print each change record. If nothing else is selected, this is the default.")
    parser.add_argument("--print-summary", "-ps", action="store_true", help="Print a summary of a change record")
    parsed = parser.parse_args()

    handlers = []
    if parsed.print_record:
        handlers.append(print_change_record)
    if parsed.print_summary:
        handlers.append(print_summary)
    
    if len(handlers) == 0:
        # When no handlers are set, we default to printing the record
        handlers.append(print_change_record)

    start_watching(parsed.stream_arn, handlers)

if __name__ == "__main__":
    main()
```

Thanks to argparse, we get this nice help menu when calling the script.

```terminal
$ python dynamodb_streamgazer.py -h
usage: dynamodb_streamgazer.py [-h] [--print-record] [--print-summary] stream_arn

See what's going on in DynamoDB Streams in near real-time 🔍

positional arguments:
  stream_arn            The ARN of the stream you want to watch.

optional arguments:
  -h, --help            show this help message and exit
  --print-record, -pr   Print each change record. If nothing else is selected, this is the default.
  --print-summary, -ps  Print a summary of a change record
```

Here's an example of the script being called with only the summary option. The delay between the changes happening in the table and the output showing up in the console is negligible. It's also straightforward to implement your own Observers that can do aggregations or other suitable things for your workflow.

```terminal
python dynamodb_streamgazer.py $STREAM_ARN --print-summary
Starting watcher for shard: shardId-00000001653646993166-46aa7561
Starting watcher for shard: shardId-00000001653648537152-e0a56e69
Starting watcher for shard: shardId-00000001653648750475-f3978e9b
Starting watcher for shard: shardId-00000001653657153330-46f0ba41
[2022-05-27T15:35:57+02:00] - INSERT - PK=test, SK=item
[2022-05-27T15:36:13+02:00] - MODIFY - PK=test, SK=item
[2022-05-27T15:36:23+02:00] - REMOVE - PK=test, SK=item
```
In this post, I've introduced you to a script that allows you to look into DynamoDB Streams in near-real-time. The [code is available on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-streamgazer). Hopefully, you find this helpful, and I'm looking forward to your feedback and questions.

---

Title Image by [Towfiqu barbhuiya](https://unsplash.com/@towfiqu999999?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/transactions?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)