------
title: "IAM Conditions - Force providing specific tags during resource creation"
author: "André Reinecke"
date: 2022-04-07
toc: false
draft: false
image: "img/2022/04/iam_forcing_tags_conditions.png"
thumbnail: "img/2022/04/iam_forcing_tags_conditions.png"
categories: ["aws"]
tags: ["level-200", "iam", "policy", "condition", "well-architected"]
summary: |
    We can use permissions boundaries to give developers and teams more freedom to create their own resources. For forcing them to provide specific tags during resource creation, we need a deeper understanding of how this can be achieved. We talk about the example of creating a security group.
---

To continue the discussion on [balancing security and developer productivity](/2022/03/using-permission-boundaries-to-balance-security-and-developer-productivity.html) by [Maurice Borgmeier](/authors/maurice-borgmeier.html), I want to give another example on how to provide the freedom to create resources, but enforce guidelines on tagging those resources.

Let's say we are using one AWS account for multiple teams in a developer stage and want them to create their own resources. But _only_ if the resources will be tagged correctly so that they afterward just have access to their own resources.

We are using the example of a [security group](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html). We can create a security group in our VPC with a single awscli-call:

```bash
aws ec2 create-security-group \
    --description "Allow traffic for my lambda-function" \
    --group-name "team1-sg-for-lambda-dev" \
    --vpc-id "vpc-abc12345"
```

For this call, we (the principal who makes the API call) need the following permissions. To further restrict the creation to a specific VPC, we want to specify a special VPC-Id:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": [
        "arn:aws:ec2:eu-central-1:123456789123:security-group/*",
        "arn:aws:ec2:eu-central-1:123456789123:vpc/vpc-abc12345"
      ]
    }
  ]
}
```

But what if you don't want to specify the VPC-Id, but tags, which belong to that specific VPC? We can enable this with a [condition](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition.html) in our [IAM-policy-statement](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_statement.html). For the condition we need to use some comparison with given tags on the VPC resource. We use a so-called _condition context key_ named [_aws:ResourceTag_](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_condition-keys.html#condition-keys-resourcetag). Why do we need to use this one? We expect that the VPC was tagged beforehand so that the existing resource (the VPC) already has tags, which we can use to compare within our policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCreatingASecurityGroup",
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:security-group/*"]
    },
    {
      "Sid": "AllowSecurityGroupCreationOnlyInVPCWithGivenTags",
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:vpc/*"],
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/Name": "DevVPC",
          "aws:ResourceTag/environment": "dev"
        }
      }
    }
  ]
}
```

Notice, that I removed the VPC id and added multiple tags which need to be present to the VPC. Also, I made two statements out of the single statement. I just want the condition to be evaluated for the VPC, in which the security group should be created.

The permission for [ec2:CreateSecurityGroup](https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonec2.html) has two possible resources: one is [security group](https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonec2.html#amazonec2-security-group) (our first statement) and the second is [vpc ](https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonec2.html#amazonec2-vpc).

Now you might ask yourself: What does this have to do with enforcing tagging to our newly created resources? Bare with me. We know how to add a security group to a given VPC with specific tags using _aws:ResourceTag_. But as of now, we don't need tags for the newly created security group. Now we talk about that.

You might have guessed it. We again need a condition. Since we don't have a resource yet, which is already tagged, we cannot use _ec2:ResourceTag_ here. We are just requesting a new resource to be created. And here we are. There is an [_aws:RequestTag_](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_condition-keys.html#condition-keys-requesttag) context key, which we need to use.

Lets add _aws:RequestTag_ to our first statement:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:security-group/*"],
      "Condition": {
        "StringLike": {
          "aws:RequestTag/Name": "team1-*-dev"
        },
        "StringEquals": {
          "aws:RequestTag/Environment": "dev"
        }
      }
    },
    {
      "Sid": "AllowSecurityGroupCreationOnlyInVPCWithGivenTags",
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:vpc/*"],
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/Name": "DevVPC",
          "aws:ResourceTag/environment": "dev"
        }
      }
    }
  ]
}
```

Of course we now need to extend our awscli call to provide some tags during creation of our security group:

```bash
aws ec2 create-security-group \
    --description "Allow traffic for my lambda-function" \
    --group-name "team1-sg-for-lambda-dev" \
    --vpc-id "vpc-abc12345" \
    --tag-specifications 'ResourceType=security-group,Tags=[{Key=Name,Value=team1-sg-for-lambda-dev},{Key=Environment,Value=dev}]'
```

If you try this, I am expecting that you will get an _UnauthorizedOperation_ error like:

```
An error occurred (UnauthorizedOperation) when calling the CreateSecurityGroup operation:
You are not authorized to perform this operation. Encoded authorization failure message: [...]
```

Why is that?

The answer is pretty simple: we don't have permission to actually tag our newly created resource. If tags are provided in our request, AWS evaluates whether the principal is permitted to create the tags (in our case _ec2:CreateTags_). [Here](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/supported-iam-actions-tagging.html), we can read that:

> To enable users to tag resources on creation, they must have permissions to use the action that creates the resource, such as `ec2:RunInstances` or `ec2:CreateVolume`. If tags are specified in the resource-creating action, Amazon performs the additional authorization on the `ec2:CreateTags` action to verify if users have permission to create tags. Therefore, users must also have explicit permissions to use the `ec2:CreateTags` action.

We don't want our developers to add or change tags on existing resources. Only during the creation of the resources are they forced to provide special tags. So let's add it to our policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:security-group/*"],
      "Condition": {
        "StringLike": {
          "aws:RequestTag/Name": "team1-*-dev"
        },
        "StringEquals": {
          "aws:RequestTag/Environment": "dev"
        }
      }
    },
    {
      "Sid": "AllowTaggingSecurityGroupsDuringCreation",
      "Effect": "Allow",
      "Action": ["ec2:CreateTags"],
      "Resource": "arn:aws:ec2:eu-central-1:123456789123:*/*",
      "Condition": {
        "StringEquals": {
          "ec2:CreateAction": "CreateSecurityGroup"
        }
      }
    },
    {
      "Sid": "AllowSecurityGroupCreationOnlyInVPCWithGivenTags",
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": ["arn:aws:ec2:eu-central-1:123456789123:vpc/*"],
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/Name": "DevVPC",
          "aws:ResourceTag/environment": "dev"
        }
      }
    }
  ]
}
```

_Be aware_ that keys and values in policies are case **in**senstive, whereas keys and values in tags are case-sensitive! You can read more about that [here](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition.html). If you need to enforce case-sensitive tags, you need to have a look at [aws:TagKeys](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_condition-keys.html#condition-keys-tagkeys).

Of course, you can add more than just the single _CreateSecurityGroup_ action to the comparison inside the conditions block. Make it a list. You might use it also for _ec2:RunInstance_ etc.

Creating the rules for our security group is out of scope for this article. This is one way how you might want to force tagging on newly created resources.

## Summary

We learned how to use tags within our IAM policies on existing resources as well as on newly created resources. This might be another way to help administrators give teams more freedom in creating their own resources within given [boundaries](/2022/03/using-permission-boundaries-to-balance-security-and-developer-productivity.html) and force them to tag their created resources.

![Architecture](/img/2022/04/iam_forcing_tags_conditions.png)

Thanks for reading. If you have further questions, you can reach out to me via the channels mentioned in my bio.

&mdash; André
