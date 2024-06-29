---
title: "The CDK Book: The missing Go Code Examples"
author: "Gernot Glawe"
date: 2022-01-09
draft: false
image: "img/2022/01/thecdkbook-go.png"
thumbnail: "img/2022/01/thecdkbook-go.png"
toc: true
keywords:
    - iac
    - go
    - cdk
    - level-300
tags:
    - level-300
    - go
    - lambda
    - cdk
categories: [aws]

---



<!--more-->

## The CDK Book

The CDK Book "A Comprehensive Guide to the AWS Cloud Development Kit" is a  book by Sathyajith Bhat, Matthew Bonig, Matt Coulter, Thorsten Hoeger written end of 2021.

Because the CDK itself is polyglott with [jsii](https://github.com/aws/jsii), the TypeScript examples are automatically translated in other languages. So the example CDK code used in the book is jsii generated, and there are samples for TypeScript, Python, Java and C#.

If you buy the book from gumroad, you get different versions for Typescript, Python, Java and C#. The content itself is mostly focused on TypeScript, which makes sense, since the CDK is mainly written in TypeScript. A few chapters for Python are also included. The testing part is typescript-only.

What is missing is example content for GO. No problem, we got this covered! We have running examples for almost every chapter.This is a part of my project "Go on AWS".


### Go Samples For the CDK Book


Whenver it is possibly I provide synth-able code snippets, that means `cdk synth` produces an output. The examples usually do not generate *useful* code, they are just learning examples.

If the snippets are to small, I provide a readme.

`*` TCB = The CDK Book
`*` GoA = Go on AWS Webseite


## TCB`*` Chapter 2: What Are Constructs?

- [S3 Bucket, CloudFront Distribution](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter2/distribution)

- [Construct MySimpleWebsite](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter2/website)

## TCB Chapter 2.2.3. Level 2 Constructs

- [Grant Function](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter2/level2)

## TCB Chapter 2.3.1. The Stack Construct


- [Stack](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter2/stack)

## TCP Chapter 4.2.2 Tasks

As an alternative: see [Chapter Tools-Task](https://www.go-on-aws.com/tools/task/)

## TCP Chapter 4.2.3 Debugging

Debugging should work out of the box with vscode:

![Debugging](/img/2022/01/vscode-debug.png)

Just start (1) Debugging.

Standard Launch Json:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}"
        }
    ]
}
```

## TCP Chapter 4.2.4 Linting

.. is included in GO. 

For formatting run : `go fmt *.go`




## TCP Chapter 5 Project Layout

- For serverless project layout: [Chapter Lambda with CDK Overview](https://www.go-on-aws.com/infrastructure-as-go/cdk-go/lambda/helloworld/overview/)

- For general project layout: [Chapter Exercise CDK-VPC](https://www.go-on-aws.com/infrastructure-as-go/cdk-go/vpc/)

Notes on files:

- tsconfig.json - not needed with GO
- package.json - the go.mod has almost the same functionality, see [Chapter Modules](https://www.go-on-aws.com/getting_started/modules/)

- .gitignore - the same functionality
- .npmignore - not needed

### Tests

- See [Chapter Serverless test Pyramid](https://www.go-on-aws.com/testing-go/integration/)
- See [Chapter Unit Test of IAC CDK GO](https://www.go-on-aws.com/testing-go/integration/unit_infra/)

### Directories

- *bin/** => main/main.go
- *lib/** => files wil be in the project root and subdirs

### Projen

As GO does not need so much helper files and handles dependencies better since 1.13, projen is not needed. Please note: projen is not CDK standard project layout.


## TCP Chapter 5.1.11. Initializing with the CLI

```bash
npx cdk init app --language=go
```
Shortform

```bash
npx cdk init app -l=go
```



## TCP Chapter 6 Custom Resources and CFN Providers

## TCP Chapter 6.1.1. Implementing custom resources using AWS CDK

- [Sources CDK](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter6/customresource/infra)
- [Sources Lambda](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter6/customresource/app)

## TCP Chapter 6.2. CloudFormation Resource Types

- [Sources](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter6/cfnresource)

## TCP Chapter 7.  Configuration Management

*simpleapiwithtestsstack*:

- [Configuration Management Sources](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter7/configmanagement)

## TCP Chapter 7.1.  Configuration Management

*DBStack*:

- [Configuration Management Sources](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter7/context/dbstack)

## TCP Chapter 7.1.  Static Files

The kind of configuration is different for the several languages.

Here an example for using Systems Manager with paddle "github.com/PaddleHQ/go-aws-ssm" and configuration files with "viper" 	"github.com/spf13/viper"

The example is a working webserver instance example. Note the *readme.md* in the source directory

- [Configuration with viper and ssm](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook/chapter7/instance)


## TCP Chapter 7.3.1. Systems Manager Parameter Store

```go
awsssm.NewStringParameter(stack, aws.String("Parameter"),
    &awsssm.StringParameterProps{
		AllowedPattern: aws.String(".*") ,
        Description:   aws.String("The value Foo"),
        ParameterName: aws.String("FooParameter"),
        StringValue:   aws.String("Foo"),
		Tier: awsssm.ParameterTier_ADVANCED,
    },
```

## TCP Chapter 8.3.1 Docker Image Asset in Action: ECS Fargate


A similar example is in the architecture section
- [Go on AWS Chapter Fargate Container](https://www.go-on-aws.com/architectures/fargate/)

- [Source for Fargate Contrainer](https://github.com/megaproaktiv/go-on-aws-source/tree/main/architectures/container)

- [Source Container in ECS pattern](https://github.com/megaproaktiv/go-on-aws-source/tree/main/architectures/container/infra/3-container/fargate.go)

## TCP 8.4.1. S3 Assets in Action: Deploying Lambda Functions using S3 Assets

A similar example is in the Lambda Construct section:

- [GoA Chapter CDK Lambda Construct](https://www.go-on-aws.com/infrastructure-as-go/cdk-go/lambda/helloworld/cdk_lambda_construct/)
- [Source CDK Lambda S3 Asset](https://github.com/megaproaktiv/go-on-aws-source/tree/main/lambda-go/lambda-cdk/hello-world)

## TCP 8.5 Docker Bundling

- [Source Docker Bundling](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/lambda/deploy_bundler/infra)

- [GoA Chapter Docker Bundling](https://www.go-on-aws.com/lambda-go/deploy/deploy_lambda_docker_bundling/)

## TCP 9 Testing

Most of the concepts are described with code examples in 
- [GoA Chapter Serverless Test Pyramid](https://www.go-on-aws.com/testing-go/integration/)
- [CDK Infrastructure Testing](https://www.go-on-aws.com/infrastructure-as-go/cdk-go/cit/)




## See the full source on github.

- [Sources](https://github.com/megaproaktiv/go-on-aws-source/tree/main/infrastructure-as-go/cdk-go/thecdkbook)


## Feedback & discussion

For discussion please contact me on twitter @megaproaktiv

## Learn more GO

Want to know more about using GOLANG on AWS? - Learn GO on AWS: [here](https://www.go-on-aws.com/)

