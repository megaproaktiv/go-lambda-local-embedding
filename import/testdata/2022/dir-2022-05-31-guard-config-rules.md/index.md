---
title: "New AWS Config Rules - LambdaLess and rust(y)"
author: "Gernot Glawe"
date: 2022-06-01
draft: false
image: "img/2022/05/guard-header.jpeg"
thumbnail: "img/2022/05/guard-header.jpeg"
toc: true
keywords:
    - security
    - config
    - rust
tags:
    - level-300
    - security

categories: [aws]

---

AWS Config checks all your resources for compliance. With 260 managed rules, it covers a lot of ground. But if you need additional checks until now, you had to write a complex Lambda function. With the new "Custom Policy" type, it is possible to use declarative Guard rules. Custom Policy rules use less lines of code and are so much easier to read. 

<!--more-->


AWS CloudFormation Guard is a "general-purpose policy-as-code evaluation tool", which means it interprets rules. The Guard tool is used inside AWS Config **and** you can use it offline , which makes development easy. It is written in [Rust](https://aws.amazon.com/sdk-for-rust/).

The first part of this post shows you the example from the AWS documentation, which checks a DynamoDB table. In the second part I will show you how to create these rules yourself.

What is AWS Config?

## AWS Config
[AWS Config](https://docs.aws.amazon.com/config/latest/developerguide/WhatIsConfig.html) checks your resources and provides configuration snapshots. All changes of these *Configuration Items* are stored in a timeline. So you may exactly see when resource configuration changes. In addition to that, you may query the config database for resources. 

What types of rules exist?

## AWS managed Rules

There are 260 managed rules. You can apply them quickly, and no code is needed. A list is provided in the [Config documentation](https://docs.aws.amazon.com/config/latest/developerguide/managed-rules-by-aws-config.html).

## Custom Lambda Rules

Sometimes you need additional or special checks. [Until April 2022](https://docs.aws.amazon.com/config/latest/developerguide/DocumentHistory.html), you had to write a custom Lambda rule. The AWS rdk - rule development kit supports the development of custom Lambda rules.
I have built some of those rules for several projects and found the development process quite complex. If you want to check only an attribute of a resource, you need about 40 lines of code.


### Checking DynamoDB with a custom Lambda rule.

For instance if you want to check a DynamoDB table, in the Lambda function the first thing you have to do is check the table status, like in the **DYNAMODB_ENCRYPTED** Rule:

 ```python
status_table = configuration_item["configuration"]["tableStatus"]
	 if status_table == "DELETING":
	 return build_evaluation_from_config_item(configuration_item, 'NOT_APPLICABLE')
```

You can see the whole 440 line Lambda function in the [AWS Config Rules repository on github](https://github.com/awslabs/aws-config-rules/blob/master/python/DYNAMODB_ENCRYPTED_CUSTOM/DYNAMODB_ENCRYPTED.py).

## Custom Policy Rules

This is the same check as Guard 2.0 rule fragment:

```sql
 when configuration.tableStatus == 'ACTIVE'
```

There is also a change in the programming logic. You do not think in programm steps, you think in rules and filters. With the Guard rules you filter the not applicable items with the `when` query. These query are documented in the [CloudFormation Guard repository](https://github.com/aws-cloudformation/cloudformation-guard/blob/main/docs/QUERY_AND_FILTERING.md).


### First complete rule: Checking DynamoDB for point-in-time backup anables

In the [Creating Custom Policy Rules](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules_cfn-guard.html) documentation the example checks a Table for point in time recovery: 

```hcl
let status = ['ACTIVE']

rule tableisactive when
    resourceType == "AWS::DynamoDB::Table" {
    configuration.tableStatus == %status
}

rule checkcompliance when
    resourceType == "AWS::DynamoDB::Table"
    tableisactive {
        let pitr = supplementaryConfiguration.ContinuousBackupsDescription.pointInTimeRecoveryDescription.pointInTimeRecoveryStatus
        %pitr == "ENABLED"
}                      
```

Let's break this down:

In the first rule, `tableisactive`,Tables are filtered for being 'ACTIVE'.

The second rule called "checkcompliance" uses a variable `pitr`(point in time recovery) to check the `supplementaryConfiguration`. 

This `supplementaryConfiguration` is the DynamoDB backup configuration. You read it with the AWS CLI:


```bash
aws dynamodb describe-continuous-backups --table-name $table
``` 

 The output is, for example:

```json
{
    "ContinuousBackupsDescription": {
        "ContinuousBackupsStatus": "ENABLED",
        "PointInTimeRecoveryDescription": {
            "PointInTimeRecoveryStatus": "ENABLED",
            "EarliestRestorableDateTime": "2022-04-25T16:38:44+02:00",
            "LatestRestorableDateTime": "2022-05-22T09:18:03.639000+02:00"
        }
    }
}
```

So `supplementaryConfiguration.ContinuousBackupsDescription.pointInTimeRecoveryDescription.pointInTimeRecoveryStatus` reads "ENABLED", which evaluates to passing the check for this example table.

Unfortunately, I have not yet found any documentation on the `supplementaryConfiguration` for different `resourceType`. 

## Development Steps for a simple example

Now for a simpler example, how to develop custom Guard rules which only checks the base configuration of the resource, not the supplementaryConfiguration.

I will take the well known Lambda Runtime example. This example is also covered as a managed rule `lambda-function-settings-check`. The check evaluates to compliant if the Lambda Function uses the newest runtime of the development languages.

## Guard Rule development steps

Step 1) Find configuration item and the structure
Step 2) Create the rule code
Step 3) Create Rule resource with debug option
Step 4) Create a PASS and a FAIL resource for testing
Step 5) Debug and change code with Logs


### Step 1: Find configuration items

The structure of the configuration item can be found with the AWS CLI describe commands, like `aws lambda get-function --function-name $name`.

```json
{
"Configuration": {
    "FunctionName": "compare-py-9",
    "FunctionArn": "arn:aws:lambda:eu-central-1:795048271754:function:compare-py-9",
    "Runtime": "python3.9"},
  "..."
  }
```

As the structure is correct, the case of the names is wrong. We need to look at the "Runtime" attribute, but is is stored as "runtime". How do I know this?
The better way is to use the Config Resource Inventory.


![Image description](/img/2022/05/7tsrracggl6mts0o53ai.png)
Youd find items in the Config Dashboard (1). Open a Lambda Function item (2).


![Image description](/img/2022/05/5auvs6brqzy0kfh8ynfb.png)
The rules are checked against this configuration item.

This is a fragment of the Config Configuration Item:

```json
{
  "version": "1.3",
  "accountId": "795048271754",
"..."
  "resourceType": "AWS::Lambda::Function",
  "resourceId": "compare-py-9",
  "resourceName": "compare-py-9",
 "..."
  "configuration": {
    "functionName": "compare-py-9",
    "functionArn": "arn:aws:lambda:eu-central-1:795048271754:function:compare-py-9",
    "runtime": "python3.9",
"..."
```

So the field we check in Config is `configuration.runtime`.

Now I create the rule. You find the documentation for Guard [in the github repository](https://github.com/aws-cloudformation/cloudformation-guard/tree/main/docs)
You will see that there are mostly CloudFormation examples, because Guard is used for CloudFormation template checks in the first place. But it can be used for any structures.

## Step 2: Create the rule code


The first part of the rule is to filter for the Resource type:

```hcl
 resourceType == "AWS::Lambda::Function" 
```

If you check at the deepest configuration level, it is a good idea to check the [API documentation](https://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunction.html#SSS-CreateFunction-request-Runtime) for the allowed values.

The allowed runtime values are:

```bash
Valid Values: nodejs | nodejs4.3 | nodejs6.10 | nodejs8.10 | nodejs10.x | nodejs12.x | nodejs14.x | nodejs16.x | java8 | java8.al2 | java11 | python2.7 | python3.6 | python3.7 | python3.8 | python3.9 | dotnetcore1.0 | dotnetcore2.0 | dotnetcore2.1 | dotnetcore3.1 | dotnet6 | nodejs4.3-edge | go1.x | ruby2.5 | ruby2.7 | provided | provided.al2
```
As this attribute is not required (Required: No) I need to check for the existence at first:

```txt
  WHEN configuration.runtime !EMPTY {
}
```

For the values itself there is the "IN" function:

```txt
configuration.runtime IN ['python3.9','go1.x','nodejs16.x']
```

So the *whole* check is 6 lines long:

```txt
rule lambdaruntimenewest when
    resourceType == "AWS::Lambda::Function" {
     WHEN configuration.runtime !EMPTY {
    configuration.runtime IN ['python3.9','go1.x','nodejs16.x']
    }
}
```

## Step 3: Create Rule resource with debug option


Open the Config Service in the AWS console and create a rule with type "Create custom rule using Guard". For this example the name is LambdaRuntime.

Be sure to add a description which tells the reader what attributes are checked.


![Image description](/img/2022/05/3n24l11u2vmbs2c0kdy5.png)

1) The Name of the rule - the CloudWatch debug log will be named accordingly, here "LambdaRuntime"
2) A descriptive description 
"Checks runtime of Lambda for newest runtime"
3) Enable logs - disable them later


![Image description](/img/2022/05/ffvdsk7al2g47zefi22n.png)
4) Rule content


![Image description](/img/2022/05/z01ocjk6kbislba1enrd.png)

5) Limit to Lambda resources
6) AWS Resources (not third party)
7) Choose Lambda 

You do not need a picture for these steps, do you?
8) Click "Next"
9) Click "Add rule" in the "Review and create" step

## Step 4 Create a PASS and a FAIL resource for testing

Now you create a Lambda Function, which passes the check. Save the configuration item as "pass.json".

Save the failing Lambda Function configuration as "fail.json"

You can use these files for local development later.

## Step 5 Debug and change code with Logs

You will find the debug log in CloudWatch Log. The log group will have the name of the rule prepended like:

`/aws/config/config-rule/LambdaRuntime/config-rule-awytio`.

Make sure to set a retention time and disable the debug output option later, because a lot of log entries are generated.


![Image description](/img/2022/05/2dn0p4mecm97xebkqwpz.png)
Now you can Re-Evaluate the new rule to trigger an execution. If your pass Function passes and your fail Function fails, you are done!


This was development with the AWS console. Now I will show you the local development. First you have to install Guard according to [Guard on github](https://github.com/aws-cloudformation/cloudformation-guard).

## Local Development

With installed Guard tool, you can develop rules locally, if the do not use supplementaryConfiguration.

Save fail configuration item as `fail.json` and passed item as `pass.json`.

If you just want to test Guard, you can use the code at [megaproaktiv on github - config-guard](https://github.com/megaproaktiv/aws-community-projects/tree/main/config-guard).


### Validate fail

```bash
cfn-guard validate --data  fail.json --rules LambdaRuntime.rule
```

The `data` parameter points to the data file and the `rules` parameter to the rules file.

Output:


```txt
fail.json Status = FAIL
FAILED rules
LambdaRuntime.rule/lambdaruntimenewest    FAIL
---
Evaluation of rules LambdaRuntime.rule against data fail.json
--
Property [/configuration/runtime] in data [fail.json] is not compliant with [LambdaRuntime.rule/lambdaruntimenewest] because provided value ["nodejs14.x"] did not match expected value ["python3.9"]. Error Message []
Property [/configuration/runtime] in data [fail.json] is not compliant with [LambdaRuntime.rule/lambdaruntimenewest] because provided value ["nodejs14.x"] did not match expected value ["go1.x"]. Error Message []
Property [/configuration/runtime] in data [fail.json] is not compliant with [LambdaRuntime.rule/lambdaruntimenewest] because provided value ["nodejs14.x"] did not match expected value ["nodejs16.x"]. Error Message []
--
```


### Validate Pass

```bash
cfn-guard validate --data  pass.json --rules LambdaRuntime.rule
```

Output:

```txt
pass.json Status = PASS
PASS rules
LambdaRuntime.rule/lambdaruntimenewest    PASS
---
Evaluation of rules LambdaRuntime.rule against data pass.json
--
Rule [LambdaRuntime.rule/lambdaruntimenewest] is compliant for template [pass.json]
--
```

## Conclusion

The new Custom Policy Rules type makes the writing of simple checks very easy. The existence of the rules engine as a local open-source tool simplifies rules development.

The supplementaryConfiguration should be documented, which is not the case now.

Also, it is the first AWS Rust backend service I have seen, so this is an exciting trend!

What do you think about rusty Custom Policy Rules?

For more AWS development stuff follow me on twitter [@megaproaktiv](https://twitter.com/megaproaktiv)

## See also 

- [Code on Github](https://github.com/megaproaktiv/aws-community-projects/tree/main/config-guard)
- [AWS Labs Lambda Config rules](https://github.com/awslabs/aws-config-rules/blob/master/python/DYNAMODB_ENCRYPTED_CUSTOM/DYNAMODB_ENCRYPTED.py)
- [ AWS Documentation Creating AWS Config Custom Policy Rules](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules_cfn-guard.html)
- [CloudFormation Guard](https://github.com/aws-cloudformation/cloudformation-guard)
