---
title: "Use the CDK to trigger your Lambda function in sub-minute intervals"
author: "Maurice Borgmeier"
date: 2022-02-01
toc: false
draft: false
image: "img/2022/02/sub_minute_title3.png"
thumbnail: "img/2022/02/sub_minute_title3.png"
categories: ["aws"]
tags: ["level-300", "cdk", "lambda", "stepfunctions"]
summary: |
    In this post I'll show you how to trigger your Lambda functions in intervals smaller than a minute using StepFunctions and the CDK.
---

Lambda functions are everywhere in AWS. One of the many use cases is to periodically perform some action. This can mean starting or stopping instances or fetching data from an API and storing it somewhere. Especially in the latter case it's useful to be able to query the API every couple of seconds. Unfortunately this isn't possible using the native CloudWatch events trigger. In this post I'll show you an alternative.

Why is this not possible natively? The common mechanism that's used to schedule Lambda functions at certain intervals or point in times is through CloudWatch Events aka. Event Bridge. The system allows you to create interval-based or cron-based rules that can notify other AWS services. Unfortunately the smallest time-resolution either of these systems support is one minute.

You could of course use the Event Bridge trigger to run the function every minute and then use `sleep()` in your code to perform the operation every couple of seconds. That has several drawbacks though. You would create a long-running Lambda function that does nothing most of the time. Also, if a single fetch-operation takes longer than your interval, your timing for the next iteration will be messed up. By this point your code wouldn't be only event based anymore, you'd effectively have to implement a scheduler in your Lambda function. Fortunately there are better options.

How can we solve this problem? We can use a combination of Event Bridge Rules and Step Function state machines to periodically trigger our Lambda functions. The diagram below shows the implementation I chose for this problem. You can think of this approach as two loops.

![Architecture](/img/2022/02/sub_minute_trigger_architecture.png)

The outer "loop" is handled by Event Bridge. Every minute a rule triggers the inner loop - a step function. This step function is set up in a way to handle the operations for a given minute. It will trigger Lambda functions asynchronously using the "Event" invocation type and then sleep until the next interval. This is a fairly simple setup and has the benefit that it's also inexpensive.

The code for this solution is written in Python and [available on Github](https://github.com/MauriceBrg/snippets/tree/main/sub-minute-lambda-trigger). In the sub_minute_lambda_trigger/infrastructure.py you'll find the following construct. It allows you to wrap a Lambda function with a construct in order to facilitate the interval trigger.

```python
class SubMinuteLambdaTrigger(Construct):

    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        interval: int,
        lambda_function: _lambda.Function,
        enabled=True,
         **kwargs
    ):
        super().__init__(scope, construct_id, **kwargs)
		# Full code on Github...
```

To use it you create an instance of the construct, pass the reference to your Lambda function to it and tell it at which interval to run the function. The rest is handled automatically by the construct. It will set up the event rule, create and configure the step function and invoke your function on time.

```python
# Create SubMinuteLambdaTrigger with Lambda Function and interval
SubMinuteLambdaTrigger(
	self,
	"sub-minute-trigger",
	interval=10,
	lambda_function=lambda_function,
	enabled=True,
)
```

There are, however, some caveats here. This solution creates a standard step function and for most purposes they will be fast/accurate enough. If you require a tighter timing, you might want to change the type of the step function to express. This will however increase your costs!

The second caveat is that this construct will only accept intervals of which 60 is a multiple, i.e. `60 % interval = 0`, so numbers like 2, 5, 10, 12, 15, 20, 30 will be accepted and numbers such as 7, 13 or anything above 30 will be rejected. This makes the whole system more accurate, because we can only rely on the outer loop triggering the state machine every minute and can't communicate an offset.

You can extend it to allow different intervals, but there is a trade off. If you want to disable this kind of trigger, you can disable the rule that triggers the step function. Given this setup, the trigger will operate for at most one more minute before it ceases invocations. If you extend the construct to allow different intervals you'd have to decrease the number of Event Bridge invocations. That means the step function will run for a longer time and it takes more time until the invocations are stopped.

## Summary

In this post I shared with you how you can use step functions and Event Bridge in combination to achieve sub-minute interval triggers for Lambda functions. This can be implemented using the CDK and [the code in this repository](https://github.com/MauriceBrg/snippets/tree/main/sub-minute-lambda-trigger).

Thank you for reading and I hope you liked this post. If there are any questions or concerns, feel free to reach out to me through the social media channels listed in my bio.

&mdash; Maurice

