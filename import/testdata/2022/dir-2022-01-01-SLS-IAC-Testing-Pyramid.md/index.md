---
title: "Views of the Pyramids: From a monolithic Test process to a Serverless Test Automation with CodeBuild"
author: "Gernot Glawe"
date: 2022-01-01
draft: false
image: "img/2021/12/testpyramid-overview.png"
thumbnail: "img/2021/12/steffen-gundermann-PtGvu2P-Gco-unsplash.jpeg"
toc: true
keywords:
    - test
    - iac
    - go
    - lambda
    - cdk
    - level-300
tags:
    - level-300
    - go
    - lambda
    - cdk
categories: [aws]

---

Comparing the development methodology of a *monolithic* program to a Serverless IAC application you will see that the power of DevOps lies in automating everything. I will show you a working example of a serverless CI pipeline with automated unit, integration and end2end test and test reports in CodeBuild. The full source is written GO, with references to Node.JS and python for the test parts.

<!--more-->

This post focuses on the testing patterns and the implementation with CodeBuild. The CDK and Lambda code is explained in detail on [go-on-aws Serverless Testing Pyramid](https://www.go-on-aws.com/testing-go/integration/).

## Monolithic development

![](/img/2021/12/mono-method.png)

I am descriping a typical development methodology of a project that I used to know (tm):

### Development

Several  developers (Dev) coded several parts of the application. Integration of the parts is done by a local Jenkis Server. The developers perform *manual* checks of their application against a test database. The do not have access to the creation and managing the configuration of the infrastructure.

The have a code repository where they store the code. For a release, they create an artifact like compiled applications, installation scripts and SQL code to change the database to the new release. This artifact is given to the Test or QA department.

### Testing

The installation scripts are executed in the test server and test database.

For a release, there is a test plan with defined test cases, which are *manually* performed.
Results of failed test cases are reported back to the developers until all test passes.

If infrastructure is changed, like the database configuration and indexing, this is performed before the test. So in a way infrastructure is tested. The *big* difference to infrastructure as code is, that a change of the server hardware is not part of any development cycle. The size is planned upfront and just has to fit. Upscaling of server hardware is only a task of the operations department.

Testing is done according to ISTQB, see some of the test types in the online [ISTQB Glossary](https://glossary.istqb.org/en/search/).

### Operations

After successfull testing the artifacts (often not the same files as used for testing, but specially generated for production) are given from the test department to the operations team.

During a maintanance window (of several hours) the script and changes are applied and some basic end2end test are performed.

This is the "old world".

Now we have a look at Test automation for a serverless application

## Testtypes for Serverless Testing - Theory

![](/img/2021/12/iac-pyramid.png)

### Unit Testing

For the application and the infrastructure, there should be automated unit testing. This can be a totally separate test. Application and infrastructure are not *integrated* yet.

### Integration Testing

Ask yourself: Which separated parts/components are now working together? If you cannot test the components individually, they are to tightly coupled together. 

![](/img/2021/12/tight-coupling.png)

Tightly-coupled component *depend* on each other and are not usable independently. So the *testability* is a hint to good system design.

![](/img/2021/12/loose-coupling.png)

If components are loosely coupled, you can test the interface. 

For a serverless application on AWS usually you have AWS Resources like storage (like S3) or databases (like dynamodb) wich are a part of the application. But these parts should be loosely coupled, so that you could exchange components and test single components. For each component of the application you have to decide whether it should be interchangeable or not. 

The integration test can be performed only on the application side, when different component of the "software" (application) part are integrated. Also the "software" is integrated with the "hardware" (infrastructure).

### End2End Testing

End2End means performing a test from the start event is as realistic as possible - this is the first end: **End**2End.
Then one path of the application is tested with variations of parameter and the result is checked against the expected results. This is the other end: End2**End**. 

Also this is a bet on how the production environment will behave. So you will only come very near to realistic production environment, 100% is almost impossible to reach.


## Testtypes for Serverless Testing - Serverless DSL application

I show you the automation on a standard DSL - **D**ynamo **S**3 **L**ambda application. In 2018 my fellow consultant Marco Tesch had the idea to define benchmarks for IaC scenarios. We take the serverless application scenario. See the [Code on Github](https://github.com/tecracer/tRick-benchmarks/tree/master/serverless-standard)

![](/img/2021/12/dsl-architecture.png)

The Use Case description:

1) User uploads object to S3 bucket
2) Bucket upload event goes to lambda
3) Lambda writes object name with timestamp to dynamoDB

As we have a Retention Policy, CDK creates a helper Lambda as well.

## Unit Testing

## Infrastructure Unit Tests

![](/img/2021/12/infra-unit.png)

Unit testing for the CDK can be the test of the generated CloudFormation. This has two main purposes:

- Test if the iac code builds correctly
- Test if the right CloudFormation is generated

### Test if the iac code builds correctly

CDK is very volatile, you have new versions each week, interfaces are changing from unstable to stable and so on. In the early days of the CDK around v. 0.35 i checked this in: [CDK - under Construction - should we use it for the next project?](https://aws-blog.de/2019/07/cdk-under-construction-should-we-use-it-for-the-next-project.html)

### Test if the right CloudFormation is generated

We are using program logic to dynamically generate CloudFormation. So we should write test.

See [go-on-aws](https://www.go-on-aws.com/testing-go/integration/unit_infra/) for a deep dive on the code.

## Application Unit Test

![](/img/2021/12/app-unit.png)

If the application is not to tighly coupled, this test type should be easy to implement. Basic business logic should not depend on AWS Services. E.g. if you use Systems Manager Parameter store or App Config for configuration, then there should be an abstraction to this configuration in your business logic.

So in the picture the grey **Infrastructure** part can be viewed seperated from the **Application** part.
Here we have only a small part of logic as an example.


See [go-on-aws](https://www.go-on-aws.com/testing-go/integration/unit_app/) for a deep dive on the code.

## Integration Testing

## Infrastructure Integration Test

![](/img/2021/12/infra-integ.png)

Here I test whether the CDK code really created AWS Resources. You could discuss whether this really can be called "integration", but it moves the test to higher levels of the test pyramid, because its closer to production.

The purpose here is not to check, whether an "AWS::Lambda::Function" Cloudformation really creates a function. This is a safety net whether I selected the correct Constructs and you also get insights about timing of the creation.

See [go-on-aws](https://www.go-on-aws.com/testing-go/integration/integ_infra/) for a deep dive on the code.


## Application Integration Test

![](/img/2021/12/app-integ.png)

In the first part I only want to test the functionality of the Lambda function running really on AWS. This also test the IAM rights of the function, some timing and configuration etc.

So I feed a test event directly into Lamba. This way the test does not depend on the right configuration of the S3 events, you just test the lambda itself.

For the output I weigh effort against benefit. Beeing a purist you should mock the DynamoDB table for lambda. Pragmatically I just check the Table itself.

See [go-on-aws](https://www.go-on-aws.com/testing-go/integration/integ_app/) for a deep dive on the code.

## End2End Testing

Now the whole chain is tested. The test puts a file directly in the bucket and checks the Table:

1) Setup (delete testitem in table)
2) Trigger end2End event (put file in S3 Bucket)
3) Test (check table entry)
4) Teardown / Cleanup (delete testitem in table)

See [go-on-aws](https://www.go-on-aws.com/testing-go/integration/integ_app/) for a deep dive on the code.

## Automate Everything

With AWS CodeBuild, you can setup a serverless CI build environment. All test can be performed by CodeBuild and you see the test results aka reports in the AWS console also. See [AWS documentation create a build project](https://docs.aws.amazon.com/codebuild/latest/userguide/create-project.html) for details. For the start it is easier to create the CodeBuild build project with the AWS console, because you have guiding wizards and the intial roles are created automatically.

The various test output formats from the several test frameworks from Python, Node.JS and GO are not natively supported. But as the Java testing library "Junit" is the de-facto standard, you can convert all types to Junit.

See [AWS Documentation Working with test reporting in AWS CodeBuild
](https://docs.aws.amazon.com/codebuild/latest/userguide/test-reporting.html) for more detail.


### Python

Pytest supports junitxml:

```bash
python -m pytest --junitxml=<test report directory>/<report filename>
```

See [AWS documentation Set up test reporting with pytest](https://docs.aws.amazon.com/codebuild/latest/userguide/test-report-pytest.html)

### GO

You can use `go-junit-report` to convert the GO standard test outputs to JUNITXML:

As an example:

```bash
go get -u github.com/jstemmer/go-junit-report
I_TEST=yes go test -v  2>&1 | go-junit-report >$CODEBUILD_SRC_DIR/report-infra-integration.xml
```

Here I control the test type with the `I_TEST` environment variable.

### Node.JS

You can use Jest with the `jest-junit`reporter.
See [AWS documenation Set up test reporting with Jest
](https://docs.aws.amazon.com/codebuild/latest/userguide/test-report-jest.html) for details.


## CodeBuild

I use the managed Ubuntu image for the Build Project:

![build image](/img/2021/12/codebuild-image.png)


The automation steps are defined in the buildspec file. This file tells CodeBuild what steps are performed.
Here i have the full example for the example with CDK and Lambda function in GO:

### Buildspec

I am not deploying anything, in the buildspec only the tests are defined.

### Install Phase

```yaml
 version: 0.2
  2
  3 phases:
  4   install:
  5     runtime-versions:
  6       golang: 1.16
  7       nodejs: 14
  8     commands:
  9       - echo Installing CDK..
 10       - npm i cdk@v2.1.0 -g
 11       - go get -u github.com/jstemmer/go-junit-report
```

We need GO and nodejs. Here you could also define the CDK version as a variable to test new CDK versions.


### Build Phase

### Unit Test for the application

```yaml
 12   build:
 13     commands:
 14       - echo Unit Testing app...
 15       - cd $CODEBUILD_SRC_DIR/architectures/serverless/app
 16       - go test -v  2>&1 | go-junit-report >$CODEBUILD_SRC_DIR/report-app.xml
```

### Unit Test for the Infrastructure


```yaml
 17       - env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/main main/main.go
 18       - chmod +x dist/main
 19       - cd dist && zip main.zip main
```

Prepare the Lambda deployment package, because the CDK looks for this.


```yaml
 20       - echo Unit Testing infra...
 21       - cd $CODEBUILD_SRC_DIR/architectures/serverless/infra
 22       - go test -v  2>&1 | go-junit-report >$CODEBUILD_SRC_DIR/report-infra.xml
```

Perform the CloudFormation unit tests.


See code on [github](https://github.com/megaproaktiv/go-on-aws-source/blob/main/architectures/serverless/infra/dsl_test.go)

### Integration Test for CDK / Infrastructure

 ```yaml
 23   post_build:
 24     commands:
 25       - echo Deploying infra
 26       - cd $CODEBUILD_SRC_DIR/architectures/serverless/infra
 27       - cdk deploy --require-approval never
 28       - echo Integration Testing infra...
 29       - export I_TEST=yes
 30       - env I_TEST=yes go test -v  2>&1 | go-junit-report >$CODEBUILD_SRC_DIR/report-infra-integration.xml
 31       - cd $CODEBUILD_SRC_DIR/architectures/serverless/app
 32       - echo Integration Testing App...
 33       - env I_TEST=yes go test -v  2>&1 | go-junit-report >$CODEBUILD_SRC_DIR/report-app-integration.xml
 34       - cd $CODEBUILD_SRC_DIR/architectures/serverless/infra
 35       - echo Destroying infra
 36       - cdk destroy -f
```

Here the whole infrastructure is built, tested and destroyed again. You have to add some CDK cli switches so that CodeBuild does not wait for you to press "Y" in lines 27 and 36.

If you want to do the same for terraform, use [terratest](https://terratest.gruntwork.io/docs/getting-started/quick-start/).


### Define Test Reports

The units test have written  test reports. Now you tell CodeBuild where to find them:

```yaml
 39 reports:
 40   gotest_reports:
 41     files:
 42       - report-app.xml
 43       - report-infra.xml
 44       - report-app-integration.xml
 45       - report-infra-integration.xml
 46     base-directory: $CODEBUILD_SRC_DIR
 47     file-format: JUNITXML
```

## Build Policies

Please note the difference for the IAM rights. Because the CDK creates the infrastructure, the role for CodeBuild must have the matching IAM policies for that.


## Testresults

In the build history you see the results of the build itself. In this example we have a failed test in the beginning. 

![](/img/2021/12/codebuild-build-history.png)

The overall status of the test shows "failed":

![](/img/2021/12/codebuild-report.png)

Looking at the test details we see that the infrastructre integration test "TestInfraLambdaExist" failed. This test checks whether the CDK *really* created a lambda function or just pretended to do so

![](/img/2021/12/codebuild-report-list.png)

So we only have 90% pass rate:

![](/img/2021/12/codebuild-report-summary.png)


### Resolving an Error

The details of the failed test shows, that the test itself does not have enough IAM access rights:

```log
 integration_test.go:18: assertion failed: error is not nil: operation error Lambda: 
 GetFunction, https response error StatusCode: 403,
  RequestID: 2fadad1e-f943-41f0-b147-2e9a805c36da, 
  api error AccessDeniedException: User: arn:aws:sts::555555555555:assumed-role/codebuild-go-on-aws-source-service-role/AWSCodeBuild-aee11252-e13c-4e4a-bda1-15f3864a87f2 
  is not authorized to perform: lambda:GetFunction on resource: 
  arn:aws:lambda:eu-central-1:555555555555:function:logincomingobject
   because no identity-based policy allows the lambda:GetFunction action: GetFunctionConfiguration should return no error
```

So I add the policy to the `codebuild-go-on-aws-source-service-role` Role and run the build/test process again:


![](/img/2021/12/2021-12-18_22-18-35.png)

## Summary & Outlook

I have shown the possibility for a small project to get a good test coverage not only on the application part to inspire you to think about different views for automated test of the applications as a whole.

In real projects you would not directly deploy this tested application to production. You would also add QS steps e.g. with the customer. 
But as you have automated all test, the speed and agility is much faster as in the monolithic setup.

To get ideas about how to shape the development phases I recommend having a look at Emily Freemans [new model](https://siliconangle.com/2021/09/29/devops-dummies-author-emily-freeman-introduces-revolutionary-model-modern-software-development-awsq3/) for software development.

Thanks for reading, I hope you got some new ideas and insights for better application quality!

## See the full source on github.

[Sources](https://github.com/megaproaktiv/go-on-aws-source/tree/main/architectures/serverless)


## Feedback & discussion

For discussion please contact me on twitter @megaproaktiv

## Learn more GO

Want to know more about using GOLANG on AWS? - Learn GO on AWS: [here](https://www.go-on-aws.com/)

