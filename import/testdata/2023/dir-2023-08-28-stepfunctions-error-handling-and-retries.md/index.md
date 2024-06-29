---
title: "Handling Errors and Retries in StepFunctions"
author: "Maurice Borgmeier"
date: 2023-08-28
toc: false
draft: false
image: "img/2023/08/raghavendra-saralaya-X5RF8GFsX4k-unsplash.jpg"
thumbnail: "img/2023/08/raghavendra-saralaya-X5RF8GFsX4k-unsplash.jpg"
categories: ["aws"]
tags: ["level-300", "stepfunctions"]
summary: |
  "_Everything fails all the time_" has been preached to us by Werner Vogels for a few years now. Every engineer working on building and maintaining systems knows this to be true. Distributed systems come with their own kind of challenges, and one of the AWS services that help deal with those is AWS Step Functions. AWS Step Functions allow you to describe workflows as JSON and will execute those workflows for you. In this blog, we'll explore what happens when things inevitably go wrong and the options the service offers to perform error handling and retries using an example application.
---

"_Everything fails all the time_" has been preached to us by Werner Vogels for a few years now. Every engineer working on building and maintaining systems knows this to be true. Distributed systems come with their own kind of challenges, and one of the AWS services that help deal with those is AWS Step Functions. AWS Step Functions allow you to describe workflows as JSON and will execute those workflows for you. In this blog, we'll explore what happens when things inevitably go wrong and the options the service offers to perform error handling and retries using an example application.

I'm not the first to write about his topic. In fact, AWS has a blog post titled "_[Handling Errors, Retries, and adding Alerting to Step Function State Machine Executions](https://aws.amazon.com/blogs/developer/handling-errors-retries-and-adding-alerting-to-step-function-state-machine-executions/)_" with a decent introduction to the topic. Since that post was published in early 2021, the features of StepFunctions have been expanded a lot. The most notable release is the integration of the AWS SDK in [late 2021](https://aws.amazon.com/blogs/aws/now-aws-step-functions-supports-200-aws-services-to-enable-easier-workflow-automation/), which allows you to make direct API calls to almost any AWS service from your state machine. Using some of those in a recent project makes me think I have something to add to the conversation, but you'll be the judge of that.

Broadly speaking, two features help you deal with errors in StepFunctions:

1. [Retries](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-error-handling.html#error-handling-retrying-after-an-error) that allow you to - you guessed it - retry a Task or API call and optionally supports patterns like [exponential backoff](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/).
2. [Error Catchers or Fallback states](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-error-handling.html#error-handling-fallback-states) act like a try...catch or try...except block around your task and allow you to transition to other tasks if a specified error occurs.

Let's see how we can use them by creating a small application. The business logic of this demo is expressed as pseudocode here:

```python
# Initial Setup
createTableIfNotExists()
createCounterStartingAt(0)

while True:
	try:
		deleteCounterIfLimitIsReached(2)
	except CounterBelowLimitError:
		incrementCounterBy(1)

# Clean up
deleteTable()
```

We could put all of that in a Lambda function, but that wouldn't be very interesting and doesn't teach us much about Step Functions. That's why I created a state machine that expresses this logic. You can find the [code for all of this on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sfn-retries-error-handling).

![Step Function State Machine](/img/2023/08/sfn_errors_and_retries_icons.png)

I chose to implement the business logic entirely in the Step Function as it's purely based on AWS API calls and a bit of control logic. If you analyze the pseudocode, you can see that the state machine needs to deal with a few corner cases:

1. The table may already exist, calling `CreateTable` to fail
2. When we try our initial `PutItem` call, the table may still be in status `CREATING` and not yet `ACTIVE`, causing our API call to fail
3. The `DeleteItem` API call should only delete the item if the counter has reached the value 2, which is enforced by a condition expression. An exception will be raised if the counter is not yet at 2.

As you can see in the diagram, which I exported from the Workflow Studio of the Step Functions service (possibly the best UI AWS has built to date), our corner cases 1.) and 3.) are handled through error catching. This is what the error catcher for 1.) looks like in the Python CDK code and the Step Function definition:

```python
# Python CDK
create_table_if_not_exists.add_catch(
	handler=put_item,
	errors=["DynamoDb.ResourceInUseException"],
)
```

```json
// State Machine Definition
"Create Table if not exists": {
  // ...
  "Catch": [
	{
	  "ErrorEquals": ["DynamoDb.ResourceInUseException"],
	  "Next": "Create Item with Counter = 0"
	}
  ],
// ...
}
```

The important part here is how the error message is spelled. The [API docs for the `CreateTable` API](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html#API_CreateTable_Errors) just specify `ResourceInUseException`. Error matching is case-sensitive, meaning you must add the service prefix in _PascalCase_ spelling to catch that specific error. Interestingly, the API call itself needs to be specified in _camelCase_ as opposed to the _PascalCase_ spelling in the docs. Unfortunately, I wasn't able to find docs explaining this or the reasoning behind it. It's just a pattern I've observed in the wild (I have yet to find _kebab-case_ and _snake_case_, though).

If you don't plan to catch any specific error message, there are also a [number of predefined error names](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-error-handling.html#error-handling-error-representation) that you can use. The catch-all one is called `States.ALL`, which is slightly misleading because it actually catches all but one (`State.DataLimitExceeded` can't be caught; it's terminal). Additionally, you can define multiple error catchers and multiple errors per catcher.

When an error is caught, the console makes that visually clear using orange color, a nice warning Icon, and the `TaskFailed` status in the event log. But wait, if the counter starts at zero, our `DeleteItem` call should have also thrown an error at some point - why is it green? You can see all the TaskFailed messages in the event log, but the last attempt succeeded. That's why the final output is green.

![AWS Console: Stepfunction Error Catching](/img/2023/08/sfn_error_catcher.png)

In this example, I've done something somewhat dangerous that could lead to an infinite loop. The "_Delete Item if Counter = 2_" step has the "_Increment Counter_" step as the error catcher, and "_Increment Counter_" has "_Delete Item if Counter = 2_" as its next step. Be careful with that in the real world; it could become expensive.

Now that we have paid a lot of attention to error catchers, it's time to move on to retries. We're going to use a retry at the "_Create Item with Counter = 0_" task in our state machine because it will be executed right after we create the DynamoDB table. That means the table may not be ready to receive items yet, so we can retry that step later. Here's what that looks like as code:

```python
# Python CDK

# This means we'll retry after 3, 6, 12, and 24 seconds. Usually,
# the table should be available by then.
put_item.add_retry(
	errors=[
		"DynamoDb.ResourceNotFoundException",  # Table not active
	],
	backoff_rate=2,
	interval=Duration.seconds(3),
	max_attempts=4,
)
```

```json
// State Machine Definition
"Create Item with Counter = 0": {
  // ...
  "Retry": [
	{
	  "ErrorEquals": ["DynamoDb.ResourceNotFoundException"],
	  "IntervalSeconds": 3,
	  "MaxAttempts": 4,
	  "BackoffRate": 2
	}
  ],
// ...
```

If the syntax reminds you of the error catcher, you're correct - the same naming patterns apply here. If you don't specify anything beyond `ErrorEquals`, the state machine will attempt to retry the task 3 times with an interval of 1 second and a backoff rate of 2.0. To disable the exponential part of exponential backoff, you just set the `BackoffRate` to 1.

In the console, tracking the number of retries for a task is possible. This time, I don't have any caveats, I just wish AWS would show us the time since the original or previous attempt as a number for each retry. The time graphic is pretty, but I need to look at the event stream to see some concrete values (#awswishlist).

![AWS Console: Stepfunction Retries](/img/2023/08/sfn_retries.png)

That's all, folks. If you want to learn more about defining state machines with AWS API calls using the CDK, I suggest you check out the [implementation on Github](https://github.com/MauriceBrg/aws-blog.de-projects/blob/master/sfn-retries-error-handling/infrastructure.py). The generated state machine in all its glory is also [available there](https://github.com/MauriceBrg/aws-blog.de-projects/blob/master/sfn-retries-error-handling/README.md#definition) should you wish to reuse parts of it.

Thank you for your time, and I hope you learned something new.

&mdash; Maurice

---

Photo by [Raghavendra Saralaya](https://unsplash.com/@numoonchld?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/photos/X5RF8GFsX4k)
