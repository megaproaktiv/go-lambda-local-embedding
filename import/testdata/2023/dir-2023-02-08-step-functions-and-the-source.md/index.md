---
title: "Step functions and the source"
author: "Patrick Schaumburg"
date: 2023-02-08
toc: true
draft: false
image: "img/2023/02/step-functions-workflow-overview.png"
thumbnail: "img/2023/02/step-functions-workflow-overview.png"
categories: ["aws"]
tags: ["stepfunctions"]
---

Possibilities of using AWS Step Functions are widespread. Most times, filtering out the necessary content is done within Lambda Functions or other services. With this blog, I will show you how to prevent this within Lambda using in- and output filters in AWS Step Functions.
<!--more-->

## Overview

Before going into detail, a short overview of the structure.

On the left side, you see the API Gateway. It is the trigger to start an execution. Besides the function as the trigger, it also contains an input used within AWS Step Functions.

Within AWS Step Functions, a State Machine executes our workflow. The workflow itself contains two Lambdas. The first Lambda requires an input from the API Gateway and creates an output. This output is then used again by the second Lambda, which also returns an output. This output is then returned to the API Gateway, where it is used.

![Infrastructure Overview](/img/2023/02/step-functions-infra-overview.jpeg)

## Basic understanding of AWS Step Functions

AWS Step Functions allows adding automated workflows that can be triggered from outside to perform actions.
Such actions can be invoking a Lambda Function, attaching an EBS volume, creating a Kinesis Video stream, and much more. The actions have a wide range.
As of now, there are two different types of an AWS Step Function:

- **Standard**
  - Can run up to a year
  - are used for ETL, ML, e-commerce, and automation (over 2,000 per second)
- **Express**
  - Can run up to 5 minutes
  - are used for IoT and all other fast executions (over 100,000 per second)
  - debugging and execution logs are only available within CloudWatch Logs

Both have their right to exist and are great when you have it implemented.

## The source is an API Gateway

In my case, I have an API Gateway, which invokes the StepFunction with the Service integration as the target of a POST command. The API Gateway also contains a mapping template within the integration request. This mapping template is required to deliver the input (it contains user information) and the information about which State Machine to use within AWS Step Functions.

So within the API Gateway, it looks like the following.

![API Gateway Source POST](/img/2023/02/step-functions-api-gateway-source.png)

**Mapping Template**: I will only go into a little detail as this is part of a SaaS I created. The SaaS uses Amazon Cognito as the authorizer within my API Gateway and delivers the username out of the $context.

```json
{
  "input": "{\"username\": \"$context.authorizer.claims['cognito:username']\"}",
  "stateMachineArn": "arn:aws:states:eu-central-1:123456789012:stateMachine:MyStateMachine"
}
```

The JSON above contains the State Machine to use and the input data for the first step of the execution. So when we want to add data to our workflow, we must add this information into the **input**. So right now, we only have the username to be used. To make it easier to follow, I have added a fictional mail address.

```json
{
  "username": "unicorn-sales@tecracer.com"
}
```

## The AWS Step Function State Machine Workflow

When using AWS Step Function you must create a new State Machine. A State Machine contains the workflow that is executed when triggered.
To visualize the workflow with my two Lambda Functions, I have opened the Workflow Studio (which is, by the way, awsome) that defines the workflow of my State Machine.

![Step Function overview](/img/2023/02/step-functions-workflow-overview.png)

Besides the workflow's, we are also using Workflow Studio for the next steps.

### Lambda 1 Input and Output configuration

Clicking on **Lambda 1** opens the _Configuration_ Tab. Please scroll down to the headline **Payload** in the now opened window.
The setting should be set to `Use state input as payload`. Enabling the option means the payload of the API Gateway will be used and not ignored. This option now uses the input with the username.

Our Lambda 1 (written in ruby) is configured to use the [event data](https://docs.aws.amazon.com/lambda/latest/dg/lambda-services.html#event-driven-invocation), which contains the username.

```ruby
def handler(event:, context:)
  username = event['username']
  ...
  { event: JSON.generate(event), context: JSON.generate(context.inspect) }
  ...
end
```

The event output with the username, enriched with some other parameters like a user, domain, etc. will then be returned by Lambda 1.

Now, as our Lambda gives us an event back, we switch to the _Output_ tab. The event data of Lambda 1 is delivered as a payload, whereas we only wanted to have the event data as it is needed for our next Lambda Function.

To filter out only the event data, scroll again until you see the checkbox for `Filter output with OutputPath`. Enable the checkbox and add the following to the input field: `$.Payload.event`

The now filtered content is passed to the next state within our workflow.

### Lambda 2 and the final output

Our **Lambda 2** is easier on the input part. Ensure you took the same setting in the **Configuration** tab as we did for Lambda 1. The input does not need to be filtered as we already made every preparation in the first workflow step for **Lambda 1**.

Lambda 2 is performing some tasks within AWS, such as creating an AppStream 2.0 streaming URL for the username passed from the API Gateway with enriched information from Lambda 1.

```js
exports.handler = (event, context, callback) => {
  ...
  const username = event.username;
  var params = {
    ...
    UserId: username,
    ...
   };
  ...
  var request = appstream.createStreamingURL(params);
  ...
  on('success', ...
    callback(null, {
        Message: url,
        Reference: awsRequestId,
      });
  }).
  ...
  send();
}
```

The payload now that is returned by Lambda 2 contains no more event data but a message and a reference for further use in API Gateway.

```json
{
  Message: url,
  Reference: awsRequestId,
}
```

Also, for Lambda 2, we want to have only some information back. So let's open the tab _Output_ again, go to `Filter output with OutputPath`and enable it. This time we are using only `$.Payload` in the input field.

When Lambda 2 has been executed, we have reached our final step. As the API Gateway was the initial trigger for the execution, the full payload will be returned to it.

## The full definition of my State Machine in Step Functions

As Workflow Studio is not the choice of everyone or you want to include this in your Terraform code, following the JSON with the mentioned configurations.

```json
{
  "Comment": "A description of my state machine",
  "StartAt": "Lambda 1",
  "States": {
    "Lambda 1": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "arn:aws:lambda:eu-central-1:123456789012:function:lambda-ruby:$LATEST"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 2,
          "MaxAttempts": 6,
          "BackoffRate": 2
        }
      ],
      "Next": "Lambda 2",
      "OutputPath": "$.Payload.event"
    },
    "Lambda 2": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "arn:aws:lambda:eu-central-1:123456789012:function:lambda-python:$LATEST"
      },
      "Retry": [
        {
          "ErrorEquals": [
            "Lambda.ServiceException",
            "Lambda.AWSLambdaException",
            "Lambda.SdkClientException",
            "Lambda.TooManyRequestsException"
          ],
          "IntervalSeconds": 2,
          "MaxAttempts": 6,
          "BackoffRate": 2
        }
      ],
      "OutputPath": "$.Payload",
      "End": true
    }
  }
}
```

In AWS CDK you would utilize `inputPath` or `outputPath` for this.

## Summary

- Use input filter to remove unnecessary code in your Lambda Function that might bring you a longer execution time or errors during maintenance.
- Use output filter to pass only the necessary contents to your next workflow step.

Thanks for reading!

&mdash; Patrick
