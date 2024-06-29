---
title: "Waiting for things to happen and paginating responses with boto3"
author: "Maurice Borgmeier"
date: 2022-06-17
toc: false
draft: false
image: "img/2022/06/markus-spiske-oxoinV_hjko-unsplash.jpg"
thumbnail: "img/2022/06/markus-spiske-oxoinV_hjko-unsplash.jpg"
categories: ["aws"]
tags: ["level-300", "boto3", "python"]
summary: In this blog, we'll look at two features in boto3/botocore that are often overlooked - Pagination and Waiters. In addition to that, we'll explore how these are implemented under the hood.
---

In this blog, we'll look at two features in boto3/botocore that are often overlooked - Pagination and Waiters. In addition to that, we'll explore how these are implemented under the hood.

The AWS SDK for Python comprises two packages: boto3 and botocore. Boto3 depends on botocore to perform its work as botocore is a low-level abstraction of the various AWS APIs. Both packages have something in common: You won't find the AWS API calls as Python code. Instead, they ship API definitions in JSON form aside from the Python code in the library. The clients and resources are instantiated from the JSON definitions at execution time. It's a neat implementation because it allows AWS to quickly add new API calls by just updating the API definitions and shipping the new version.

The drawback is that you have to [jump through some hoops to get Autocomplete](https://dev.to/aws-builders/enable-autocomplete-for-boto3-in-vscode-3jb4) to work and the initial instantiation of a client or resource is [comparatively costly](https://aws-blog.de/2021/02/how-boto3-impacts-the-cold-start-times-of-your-lambda-functions.html), because it requires I/O operations and parsing of JSON data.

### Pagination

It's common for API calls that can return a lot of data to employ some form of pagination. In practical terms, that means an API call produces the first `n` results and some form of continuation token that you have to include in a successive API call until you get all results. You could implement something like this using a recursive function, but you'd have to maintain the implementation. The bigger problem would probably be that you'd need to explain recursion to your colleagues, which is challenging.

Fortunately, boto3 comes with a solution for this. It supports paginators for many services and API calls that make the implementation look much more Pythonic and easier to understand. Below is an example of using a paginator to list all EBS snapshots in my account.

```python
import boto3

def ec2_list_snapshots():

    client = boto3.client("ec2")
    paginator = client.get_paginator("describe_snapshots")

    all_snapshots = []
    for snapshots in paginator.paginate():
        all_snapshots += snapshots["Snapshots"]

    return all_snapshots
```

Here, `paginator.paginate()` [returns](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/paginators.html) an iterable that the loop can use to process the partial responses. I'm just adding them to a list that contains all snapshots. Still, I could also process the subset of snapshots immediately to reduce my application's memory footprint. That depends on your use case, though.

The available paginators are all defined as JSON in the depths of the botocore package. Here, you can see the implementation ([Github](https://github.com/boto/botocore/blob/develop/botocore/data/ec2/2016-11-15/paginators-1.json)) of the `describe_snapshots` paginator:

```json
//...
    "DescribeSnapshots": {
      "input_token": "NextToken",
      "output_token": "NextToken",
      "limit_key": "MaxResults",
      "result_key": "Snapshots"
    },
//...
```

This lets us get a glimpse into how this mechanism is implemented. It knows that it can paginate the *DescribeSnapshots* API, and here it's configured what to look for in the service response to facilitate that. If the output of the API call contains the *output_token* key, it will use the value from there as the *input_token* in the subsequent API call. The results it should return are under the *Snapshots* key in the response, and for each page, in the pagination, it can limit the number of responses using the value belonging to *limit_key*. I like this generic way of implementing this pattern for different API calls that, unfortunately, sometimes use other names for the continuation token concept. We can use the [PaginationConfig](https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/ec2.html#EC2.Paginator.DescribeSnapshots) to customize how many items should be returned per page, and that will be translated to API calls with the *limit_key*.

Unfortunately, pagination is not available for the resource-based APIs from boto3. The intention is most likely that these higher-level constructs should allow for pagination themselves, but some, such as the Query operation in the DynamoDB Table resource, don't. That means using the higher-level abstraction forces you to take care of that low-level detail again - not an ideal user experience, in my opinion.

### Waiting for things to happen

API operations that perform complex control plane operations are often asynchronous, triggering the process and reporting that it has been started. An example is creating a DynamoDB table. When you call CreateTable, the API [returns](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html#API_CreateTable_ResponseElements) a table description data structure with the [TableStatus](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_TableDescription.html) field that will have the value `CREATING`. This means the API call is done, but the table can't be used yet. We have to wait until it becomes `ACTIVE`. If we want to insert items immediately, we could call the DescribeTable API in a loop until the table is active. This adds logic to our code which we'd need to test. 

The preferred way is to use waiters - not of the food-serving kind. When you look at the boto3 docs for any service, you will most likely see a waiters section. You can [find the one for DynamoDB here](https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/dynamodb.html#waiters). They're easy to use and exist for many asynchronous operations. Here's an example of code that creates a DynamoDB table and then optionally waits until that table exists.

```python
import boto3

def create_table(table_name: str, wait_until_available:bool=True):
    client = boto3.client("dynamodb")

    client.create_table(
        AttributeDefinitions=[
            {"AttributeName": "PK", "AttributeType": "S"},
            {"AttributeName": "SK", "AttributeType": "S"}
        ],
        TableName=table_name,
        KeySchema=[
            {"AttributeName": "PK", "KeyType": "HASH"},
            {"AttributeName": "SK", "KeyType": "RANGE"}
        ],
        BillingMode="PAY_PER_REQUEST"
    )

    if wait_until_available:
        waiter = client.get_waiter("table_exists")
        waiter.wait(
            TableName=table_name
        )
```

These waiters handle what we'd do as well. They periodically call the DescribeTable API to determine the current status and stop doing that once the table is active. There is an upper limit for a timeout, which you can influence through the configuration. You can tell the waiter how many times to poll the status and how long to wait between calls. By default, this queries the API every 20 seconds 25 times.

Waiters exist as JSON specifications as well. You can find the code for the waiter mentioned above in botocore ([Github](https://github.com/boto/botocore/blob/develop/botocore/data/dynamodb/2012-08-10/waiters-2.json)), and it looks like this:

```json
{
  "version": 2,
  "waiters": {
    "TableExists": {
      "delay": 20,
      "operation": "DescribeTable",
      "maxAttempts": 25,
      "acceptors": [
        {
          "expected": "ACTIVE",
          "matcher": "path",
          "state": "success",
          "argument": "Table.TableStatus"
        },
        {
          "expected": "ResourceNotFoundException",
          "matcher": "error",
          "state": "retry"
        }
      ]
    },
    "TableNotExists": {
    // ...
    }
}
```

I like how you can imagine what the Python implementation that uses this looks like from this structure. If there are no waiters for what you need, you can create an [Issue in the AWS SDK repository](https://github.com/aws/aws-sdk/issues) because the service teams provide those. [I tried that](https://github.com/aws/aws-sdk/issues/273) for DynamoDB Streams, and I'm curious to see how long it will take them to add that.

That's it for today. Hopefully, you learned a little bit about boto3 and are inspired to dive deeper into the implementation of waiters and pagination. If you have any questions, feedback, or concerns, please contact me via the channels in my bio.

&mdash; Maurice

Photo by [Markus Spiske](https://unsplash.com/@markusspiske) on [Unsplash](https://unsplash.com/s/photos/hood-car)