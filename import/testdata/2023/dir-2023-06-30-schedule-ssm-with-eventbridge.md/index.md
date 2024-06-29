---
author: "Thomas Heinen"
title: "Replace Local Cronjobs with EventBridge/SSM"
date: 2023-06-30
image: "img/2023/06/dreamstudio-clocks.png"
thumbnail: "img/2023/06/dreamstudio-clocks.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "ssm", "eventbridge", "iac", "terraform"]
---
Every machine has recurring tasks. Backups, updates, runs of configuration management software like Chef, small scripts, ...

But one of the problems in a cloud environment is visibility. Instead of scheduling dozens of cron jobs or tasks per instance, would it not be nice to have a central service for this?

You already have. And it's called EventBridge...

<!--more-->

## CloudWatch Events, EventBridge, what?

A long time back, scheduling and event-driven workflows were part of CloudWatch under the name "CloudWatch Events". Since then, it has been heavily extended and renamed into "EventBridge". If you see Terraform code etc, you might still encounter the old name - so do not be surprised.

## EventBridge modes

Without trying to duplicate the [official AWS documentation on EventBridge](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-what-is.html), I want to quickly introduce some of the concepts.

First, the Scheduler allows recurring execution of events - either in a certain interval or at defined points in time. The `rate()` syntax makes it easy to execute regular runs every few minutes or hours. On the other hand, the `cron()` method allows complex statements to address tasks for every first Friday of a month etc.

We also have the purely event-driven mode of EventBridge, which is hidden in the "Rules" section. Here, you can specify any input events (queues, notifications, even notifications from external AWS partners like Stripe or GitHub) and connect them with any AWS service. This is helpful if you want to immediately react to things like expiring ACM certificates or CloudTrail alerts.

In addition, there are a multitude of additional features like custom Event Busses, Pipes, etc. If you require some serverless event processing - EventBridge is your friend.

## Systems Manager

For our context, two of the numerous Systems Manager (SSM) features are relevant.

Most importantly, Run Documents. There are canned documents that can take parameters and then execute commands via SSM Agent. While many are already provided by AWS, you can of course create your own and use this for custom automation.

It is worth noting, that there is an SSM-integrated option to run actions periodically. The State Manager is a way to associate instances with tasks and a time to execute them. While this is indeed a big overlap with EventBridge's scheduler, it is also limited to EC2 instances. Still, it offers the same `rate`/`cron` variety for execution.

In contrast to this, EventBridge can also work cross-account (with a custom event bus), archive and replay events at a later time.

## Wiring Up the Services

One of the base concepts of regular execution via on-instance Cron or Scheduled Tasks is the specific point in time. To carry this property over, we need to set the "Flexible Time Window" option to `off`.

When you look at the AWS Web Console, you find multiple event targets predefined. Lambda, Step Functions, ECS Tasks, SQS, and others are already ready for selection. But SSM is sadly missing.

The solution for this is to choose the general API integration, which provides access to all AWS APIs inside EventBridge. You simply select the service ("Systems Manager") and the desired action ("SendCommand").

Now we come to the point where some API knowledge is required because the next field will simply expect a JSON. As this is a generic integration, the data schema of the specific command will match their API exactly.

For our `SendCommand` example, a quick search of "SSM SendCommand API" will lead us to the [`SendCommand` official API documentation](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_SendCommand.html). The trick is, to just use the properties marked as "required" plus the ones which you need.

For our example, we might end up with a JSON like this:

```json
{
  "DocumentName": "AWS-ApplyChefRecipes",
  "InstanceIds": ["i-123456789abcdef"],
  "Parameters": {
    "SourceType": "S3",
    "SourceInfo": "https://examplebucket.s3.amazonaws.com/my-cookbook.tgz",
    "RunList": "recipe[my-cookbook]"
  }
}
```

Of course, every SSM command has its specific parameters which you can revisit in the Systems Manager console under "Documents" (right at the bottom).

## Terraform Example

In a minimalistic example, this will schedule the same command we have used above.

```hcl
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
}

resource "aws_scheduler_schedule" "execute_chef" {
  name       = "execute_chef"

  flexible_time_window {
    mode = "OFF"
  }

  # schedule_expression = "rate(60 minutes)"

  schedule_expression = "cron(30 8 * * ? *)"
  schedule-expression-timezone "America/New_York"

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:ssm:sendCommand"
    role_arn = aws_iam_role.execute_chef.arn

    input = jsonencode({
      DocumentName = "AWS-ApplyChefRecipes"
      InstanceIds  = ["i-123456789abcdef"]
      Parameters = {
        SourceType = "S3"
        SourceInfo = "https://examplebucket.s3.amazonaws.com/my-cookbook.tgz"
        RunList    = "recipe[my-cookbook]"
      }
    })
  }
}

resource "aws_iam_role" "execute_chef" {
  name        = "execute_chef"
  description = "Role for Scheduling"

  inline_policy {
    name = "InlinePolicy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Sid      = "InvokeCommand"
          Effect   = "Allow"
          Action   = "ssm:SendCommand"
          Resource = "*"
        }
      ]
    })
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "scheduler.amazonaws.com"
        ]
      }
    }]
  })
}
```

{{% notice note %}}
Note: Adjust this for Least Privilege, using wildcards is not good practice in production.
{{% /notice %}}

## Summary

Once you know about this way of scheduling and understand how to work with the native APIs, things get very easy. Of course, ideally, you will use this with Infrastructure-as-Code to have repeatable deployments.
