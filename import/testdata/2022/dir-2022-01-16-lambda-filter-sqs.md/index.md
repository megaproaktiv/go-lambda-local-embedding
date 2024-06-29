---
title: "Lambda SQS Event Filters may delete your messages if you're not careful"
author: "Maurice Borgmeier"
date: 2022-01-18
toc: false
draft: false
image: "img/2022/01/lambda_sqs_filter_title.png"
thumbnail: "img/2022/01/lambda_sqs_filter_title.png"
categories: ["aws"]
tags: ["level-300", "lambda", "sqs"]
summary: |
    Lambda Event filters are a great addition to our serverless toolbox and allow us to both simplify our code as well as save money. That's great, but they can also delete messages from your SQS-queues if you're note careful. In this post I'm going to show you what to watch out for.
---

In a [recent post](https://aws-blog.de/2022/01/simplify-your-code-and-save-money-with-lambda-event-filters.html) I wrote about Lambda Event Filters and their benefits. That was a follow-up to a talk I had given on the subject internally at [tecRacer](https://tecracer.de/). During that talk, my colleague [Sebastian Möhn](https://aws-blog.de/authors/sebastian-moehn.html) asked an interesting question about the filters, which I'll look into today. He wanted to know what happens to SQS messages that don't match the filters. 

We'll get to that in a bit, but first, let's take a step back and get a brief primer on Lambda Event Filters. They were introduced in late 2021 and allow you to filter events from SQS, Kinesis or DynamoDB _before_ they trigger your Lambda function. Previously all events from those sources would trigger Lambda and you had to filter in your own code, which added complexity and cost, because the function was invoked even for events it didn't care about. With event filters that changes. We can now offload the undifferentiated heavy lifting of event filtering to AWS at no additional cost. 

It's an interesting question how the system responds to messages from SQS that don't match the filter, because SQS is different from the other two event sources that support the filter. DynamoDB Streams and Kinesis both deal with streaming data. It's part of that paradigm that multiple parties can read the same datapoint from the stream. That means it doesn't matter to other clients if one of the clients isn't interested in all the messages.

SQS on the other hand implements the producer-consumer pattern. That means one or more producers create messages that can be processed by one or more consumers. Messages are _usually_ only processed by one consumer in the pool of consumers, so in most cases it behaves like one-to-one messaging. Additionally consumers have no direct way of picking the messages they want to work on, they have to process whatever SQS gives them.

![Stream vs. Queue](/img/2022/01/lambda_sqs_stream_vs_queue.png)

This has implications for event filters between SQS-Queues and Lambda functions. If the filter rejects an incoming message from SQS, one of two things could happen:

1. The message is considered **processed** and deleted from the Queue, preventing other consumers from reading it.
2. The message is considered **unprocessed** and not deleted from the Queue. It becomes available to (other) consumers again after the visibility timeout expires.

In the first case this has the implication that messages are removed from the queue, even though other consumers may be able to process them. That means one Lambda consumer on the Queue, that doesn't want a message, has side effects for all other consumers. Depending on your architecture that may be a problem.

The second case could lead to internal problems at AWS if we assume that Lambda under the hood uses the same SQS-API that's also publicly available. To understand why, you need to know how Lambda integrates with SQS. When you set up a SQS trigger for a Lambda function, the Lambda service will periodically poll the queue for work and invoke your code whenever there are messages to process. That happens in the background and you don't really need to worry about it.

Assuming that Lambda uses the same APIs we do, there is no way for the service to skip already seen and ignored messages. So if there is only one Lambda consumer on the queue and discarded messages stay in the queue, we may experience congestion over time. That would be a less than ideal design, so my guess is that in reality the Lambda service really deletes messages that don't match the filter.

Let's now see what the docs have to say on the matter:

> You can define up to five different filters for a single event source. If an event satisfies any one of these five filters, Lambda sends the event to your function. Otherwise, Lambda discards the event.
>
> &mdash; [Lambda Event Filtering, 15.01.2022](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventfiltering.html)

I'm not a native speaker, but to me "_discards_" is a bit ambiguous. It could either mean that Lambda ignores the message and it's put back into the queue or that the service also removes the event from the queue. I'm not a big fan of ambiguity in technical documentation. Let's see for ourselves what happens.

We're going to use the following test setup, [the code is available on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/cdk-lambda-filter-sqs). A CDK-App deploys a SQS-Queue and a Lambda function that gets triggered by messages in the Queue. The event source will be filtered to only accept messages that have a `process_with_lambda` attribute with the value 1. The lambda will then increment a counter in a DynamoDB table that counts how many messages have been processed.

![Architecture](/img/2022/01/lambda_sqs_filter_architecture.png)

We'll also have a script that sends alternating messages to the Queue (`job_generator.py`). The first message will have the `process_with_lambda` attribute and the second won't. This script will send 20 messages of each kind to the queue and then we'll wait for a few seconds. If the processing worked, we should see a counter value of 20 in the table. Then we'll check how many messages are still in the queue.

If that number is 0, we know that Lambda acts according to 1) and "discarded" implies the messages are deleted from the queue. If there is a non-zero number in the Queue, Lambda sends messages back to the queue for further processing. Let's run the code:

```terminal
$ python job_generator.py
Removing summary item from table
Purging Queue
Purging queues can take up to 60 seconds, waiting...
Sending Message Group 1 with 2 messages
Sending Message Group 2 with 2 messages
Sending Message Group 3 with 2 messages
[...]
Sending Message Group 18 with 2 messages
Sending Message Group 19 with 2 messages
Sending Message Group 20 with 2 messages
```

The script first resets the counter in the table and purges the already existing messages from the Queue. Since that can take up to 60s, I've added a sleep for one minute here. Afterwards we send 40 messages in total to the Queue. Then it's time to wait for a few seconds to give Lambda time to process the messages asynchronously. Afterwards we can run the script to fetch the results:

```terminal
$ python get_result.py   
Lambda processed 20 records
The Queue contained 0 messages
```

As expected the Lambda function received the messages according to the filter. It turns out that option 1) is what AWS implemented when it comes to dealing with messages that don't match the filter. Messages that don't match it are simply deleted from the Queue.

## Conclusion

We learned that the Lambda Event Filter deletes messages from the Queue when they don't match the filter criteria. That has not been completely clear from the documentation, but its a sensible implementation given the constraints imposed by the SQS API.

This implementation has implications for queues that have multiple consumers of messages. If you add a Lambda function with an event filter to the set of consumers, there might be data loss, if the other subscribers could process messages not intended for Lambda. In that case you'll lose messages without a clear indication why.

![Lambda Filter SQS delete](/img/2022/01/lambda_filter_sqs_delete.png)

As a workaround, you could use a SNS topic in front of the Queue to fan out messages to multiple interested parties by setting up a Queue for each party. This will come with additional costs though.

My wish for the #awswishlist is that the documentation clears this up and maybe a warning is added to the AWS Console when you're working with these filters. If you know it, you can plan for it - if you don't you'll experience weird problems.

I hope you enjoyed the post and for any feedback, questions or concerns, feel free to reach out via the social media channels in my bio.


