---
title: "Prepopulate Lambda Console Testevents without dirty manual  work using Terraform"
author: "Gernot Glawe"
date: 2022-08-12
draft: false
image: "img/2022/08/manualwork.png"
thumbnail: "img/2022/08/manualwork.png"
toc: true
keywords:
    - serverless
    - terraform
    - eventbridge
tags:
    - level-300
    - terraform
    - eventbridge
    - lambda
    - serverless

categories: [aws]

---

You like Lambda testevents? Great! But with "automate everything", manual console clicks are considered dirty! Keep your hand clean by automating the creation of Lambda test events. So you can give your team, and yourself prepopulated test events. This example shows you the terraform code - because this is the fastest way. With a little effort, you can translate it to CloudFormation or AWS-CDK!


<!--more-->


 ## New Lambda feature

The Lambda console has a new feature called "Shareable test events". These "STE" allows you to share the Lambda test events defined in the console with other users in the same account. This feature is only accessible via the AWS console, so there is no automation at first glance...

 ## Lack of automation

 But as a fan of automation, you want to create your Lambda Resource with Infrastructure as Code, e.g. [terraform](https://www.terraform.io/) or [AWS-CDK V2](https://docs.aws.amazon.com/cdk/v2/guide/reference.html).

 ## Whats the documentation is not saying

The [AWS documentation "Testing Lambda functions in the console"](https://docs.aws.amazon.com/lambda/latest/dg/testing-functions.html?icmpid=docs_lambda_rss) tells you that the shareable test events are stored as a schema in the Amazon EventBridge (CloudWatch Events) schema registry named lambda-testevent-schemas.

What the documenation does not tell you is *how* the shareable test events are stored. 

That is relatively easy. When you have a Lambda function called "testee", the events are stored as `_testee-schema`.

In terraform you can create these schema and the example events.

Define the Lambda function name as variable:

```terraform
locals {
    lambda_function_name = "testee"
}
```

Create the `aws_schemas_schema`:

```terraform
resource "aws_schemas_schema" "testee" {
name          = "_${local.lambda_function_name}-schema"
  registry_name = "lambda-testevent-schemas"
  type          = "OpenApi3"
  description   = "console tests test"
```  

## A complete IaC solution: `main.tf`

You may use the file from [github](https://raw.githubusercontent.com/megaproaktiv/aws-community-projects/main/lambda-testevents/main.tf).

This is the complete code:

```terraform
locals {
    lambda_function_name = "testee"
}


resource "aws_schemas_schema" "testee" {
name          = "_${local.lambda_function_name}-schema"
  registry_name = "lambda-testevent-schemas"
  type          = "OpenApi3"
  description   = "console tests test"

  content = jsonencode(  {
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "title": "Event"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Event": {
        "type": "object",
        "required": [
          "key1"
        ],
        "properties": {
          "key1": {
            "type": "string"
          }
        }
      }
    },
    "examples": {
      "Parameter1": {
        "value": {
          "key1": "value1"
        }
      },
      "Parameter2": {
        "value": {
          "key1": "value2"
        }
      }
    }
  }
} )
}
```

## Part 1: Schema Definition

In the first part:

```terraform      
"Event": {
        "type": "object",
        "required": [
          "key1"
        ],
        "properties": {
          "key1": {
            "type": "string"
          }
        }
      }
    },
```

The schema is defined, so we only got one attribute, named "key1" with the type "string".
This definition is a little bit tricky. 

But there is a workaround to create the schema:
When you define a shared test event in the Lambda console and save this event, AWS will create the definition in the EventBridge registry for you.

## Part 2: Sample events

```terraform
"examples": {
    "Parameter1": {
    "value": {
        "key1": "value1"
    }
},
```

With the defined schema, you create examples with the matching values, so `key1` must be present to fulfill the defined schema.

The test events are called `Parameter1` and `Parameter2`. These are the names which the Lambda console will show.


 ## How to pre-populate

 So we need only a few things to pre-populate the shareable test events:

 1) A Lambda function
 1) An event scheme in the `lambda-testevent-schemas` schema registry
 1) Example events

 ## Walkthrough overview

 1) Create lambda or us existing
 1) Update content in `main.tf`
 1) Call `terraform apply`
 1) Use it

 
### Step 1 - Create Lambda

Author any lambda function from scratch:

 ![](/img/2022/08/create-lambda.png)


 ![](/img/2022/08/no-test-events.png)

 You will see no shareable test events  ![](/img/icons/1.png)in the new created function.


![Event registry](/img/2022/08/no-schema.png)

Also there is no schema in the event registry. You reach the registry:

- Go to  ![](/img/icons/1.png) EventBridge
- Choose ![](/img/icons/2.png) Schemas
- See the schemas  ![](/img/icons/3.png), none at the moment


 ### Step 2 - Update content in `main.tf`: Set Lambda name

```terraform
1 locals {
2     lambda_function_name = "testee"
3 }
```  

### Step 3 - Call `terraform apply`

If you not have done so, first do `terraform init`

```bash
terraform apply
```

Your output should look like this at start:

```log
Terraform will perform the following actions:

  # aws_schemas_schema.testee will be created
  + resource "aws_schemas_schema" "testee" {...
```

And like this when its finished.

```log
aws_schemas_schema.testee: Creating...
aws_schemas_schema.testee: Creation complete after 0s [id=_testee-schema/lambda-testevent-schemas]
```

Notice, that this task took *under* one second.

## See udpdated schema in EventBridge

![](/img/2022/08/new-schema.png)
Now the new testevents are created as `_testee-schema` ![](/img/icons/3.png).


## See testevent

And you Lambda function got the events too:

![](/img/2022/08/new-test-event.png)

- Go to  Lambda service
- Choose "Edit saved event" ![](/img/icons/1.png) in the *Test* section
- See the different events ![](/img/icons/2.png) 
- See the content of the event  ![](/img/icons/3.png)

See json in `main.tf`


```terraform
33     "examples": {
34       "Parameter1": {
35         "value": {
36           "key1": "value1"
37         }
38       },
39       "Parameter2": {
40         "value": {
41           "key1": "value2"
42         }
43       }
44     }
``` 


## Add testevents

Now you can add more test-events in the `main.tf` file and do `terraform apply` again.

```terraform
 33     "examples": {
 34       "Parameter1": {
 35         "value": {
 36           "key1": "value1"
 37         }
 38       },
 39       "Parameter2": {
 40         "value": {
 41           "key1": "value2"
 42         }
 43       },
 44       "Parameter3": {
 45         "value": {
 46           "key1": "value3"
 47         }
 48       }
 49     }
 ```

 ![](/img/2022/08/add-test-event.png)

The `Parameter3` will show in the Lambda console at once.

## Conclusion

It is easy to pre-populate Lambda test events. What do you think - does this support your development workflow? Or do you prefer local testing?

For more AWS development stuff, follow me on twitter [@megaproaktiv](https://twitter.com/megaproaktiv)

## See also 

 [Testing Lambda functions in the console](https://docs.aws.amazon.com/lambda/latest/dg/testing-functions.html?icmpid=docs_lambda_rss)
- [Terraform code](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-testevents)
