---
title: "Test driven development with AWS and golang"
author: "Gernot Glawe"
date: 2020-11-08
draft: false
image: "img/2020/11/2020-11-09_00137.jpg"
thumbnail: "img/2020/11/2020-11-09_00137.jpg"
toc: true
keywords:
    - go
    - sdk
    - testing
    - level-300
tags:
    - level-300
    - go
    - testing
categories: [aws,letsbuild]
---

## Why Go?

Go(lang) is a **fast** **strongly typed** language, which is a good fit for AWS lambda and other backend purposes. I am going to highlight some nice go features. Usually this leads to heated discussions about the "best" programming language...

<!--more-->

You do not have to read, you may watch the video:	

{{< youtube kgkcof1-zY8 >}}

For the heated discussion i will now compare go  vs. node/typescript.

### Strongly typed

Some developer favour loosely typed scripting languages like `node` or `python`. In my experience in the long run, the quality of strongly types code is better. This is especially the case when you work with the AWS events.

Let me give you an example.

Examine this json which shows an event:

```json
{
    "Stacks": [
        {
            "StackId": "arn:aws:cloudformation:eu-central-1:012345678912:stack/AwsTestJenkinsStack/1ff02980-1d0f-11eb-91c3-0a7e60142948",
            "StackName": "AwsTestJenkinsStack",
			...
		}
	]
}
```

If you do not have the event tree as a structure in your language, you have to reference via strings, like in python:

```python
stackName = event['Stacks'][0]['StackName']
```

If you have a typo e.g. in "Staks", you will *not* notice it until an error occurs. You could say that you will check that with a unit test. But this is unnecessary effort and also to late.

With go you will not be able to compile that code. You will get an error like:

```go
./counter.go:35:20: event.Staks 
undefined (type *cloudformation.DescribeStacksOutput has no field or method Staks)
```

You notice that with the error the name of the structure `DescribeStacksOutput`, which describes the schema of the json data is shown.
This json data ist the `Output` of the `DescribeStacks` api call.

With node you could/should use typescript on top of node to get a strongly typed language. 
TypeScript gives you a way out of typed variables with the `any` type, so you are not *forced* to be strongly typed. And the compilation step needed to go from typescript to javascript  leads to the second topic: Speed
	
### Speed

Go compiling is fast.
With small programs, you will not even notice that go is compiling. To compare different languages for lambda, there is a repository from tecracer:

[tecRacer Trick](https://github.com/tecracer/tRick-benchmarks/tree/master/serverless-standard)

The `cdk-lambda` directory contains a typescripted Lambda function, the `cdk-lambda-go` a golang function. They both fulfil the same purpose: Read an event and write it to a dynamodb database.

I compare the compilation speed:

This is the tsc compile time for the typescript lambda:

```bash
time npm run build ...  
7,89s user 0,51s system 153% cpu 5,459 total
```

This is the `go build` time for the go lambda:

```bash
time  go build ...
0,96s user 0,85s system 209% cpu 0,862 total
```

Both lambda functions have the same functionality. OK, this is not a statistical relevant measurement with only one sample. But when you compile several times during development you notice if you need more than **5 seconds** each or less than **1 second**. In the video, i am also experiencing the difference between an `npm install` vs a `go mod tidy`. Both commands download the needed libraries. Forty-nine packages for node, three for golang.

Speed maters even more if you have unit test which you run often. And - you really should have unit test. Which leads to the third main topic: Testing

### Testing

In some languages there are several testing frameworks, but testing is not build into the language. 
The base for testing in go is build into the language. You can add testing frameworks, but the command to start the test is:

```bash
go test
```

This will compile and call all test functions in the packages `modulename_test`, if the name of the function start with `Test`. You don't have to tag your test functions or configure them -  it's just done by convention.  

If you have a file `counter.go` with a package name `package letsbuild13`, you write a test in file `counter_test.go` with `package letsbuild13_test`.

For example this function wil be executed as test:

```go
func TestCountStacks(t *testing.T) {
...
}
```

Which leads to another topic: simplicity

### Simplicity

Go has many conventions which make coding simple. I have to admit that in the first phase of learning go you have to get used to it. But after a certain time it becomes really easy.

Two examples:

#### Exported functions 

Functions which start with a capital letter are exported

In the code : 

```go
func Count(client CounterInterface) (int){
...}
```
The naming is all what it takes to export a function.

####  Interfaces

A type which implements all functions of an interface is implementing it. 
No further "implements yxz needed". More on this later. Lets say you are convinced by now and want to give it a go - pun intented.

How to start with the AWS api and test-driven-development from the beginning?

Lets have a **walkthrough**. And you don't have to wait until there is a movie to the "book"/post. 
You may watch the video to this example above.


## (very simple) Story

I want to count the numbers of CloudFormation stacks which are active in this account. The programm should be developed with the test driven method.

### Architecture

To have a testable architecture I define an interface `CounterInterface`.

```go
type CounterInterface interface {
	DescribeStacks(...)) (*cloudformation.DescribeStacksOutput, error)
}
```

As mentioned before, all classes which have all functions defined in the interface implement it.
In this case only `DescribeStacks` is needed. The CloudFormation client from AWS SDK for go surely has it.

To create a test, I code a test-class which also has the function `DescribeStacks`.

```go
type CounterInterfaceMock struct {
	DescribeStacksFunc func(...) (*cloudformation.DescribeStacksOutput, error)
}
```	

![Interface](/img/2020/11/stackcount-interface.png)

In the simple business logic part with AWS api calls, I do not create the AWS client in the function. I just pass the client as a parameter to the function:

```go
func Count(client CounterInterface)
```

When testing - this client is a mocked client. In real life I pass an AWS CloudFormation api client.

Now, let's build!


## Starting with "main.go"

A minimal go code is like:

```go
package main

func main(){
}
```

For small scripts it is tempting to put all code into in file/module, because it is so easy. Resist!

Splitting functions into smaller modules enhanced testability, reusability and gives good karma. A main function works only as an controller and should countain no programm logic. With this approach it is possible to build code without really calling AWS services. When you develop lamba, you don't have extra deploy times, only fast and modular test.

## Getting test input data

I will code faster without calling the AWS API all the time. You just fetch a real world `DescribeCloudformation` output with the AWS cli:

```bash
aws cloudformation describe-stacks
```

The output starts like:

```json
{
    "Stacks": [
        {
            "StackId": "arn:aws:cloudformation:eu-central-1:012345678912:stack/amplify-trainerportal-dev-90853-authtrainerportal0a4ecb86-1DZBYAP6LDL7F/404f2090-0bd1-11eb-af8e-0a3f04c080ce",
            "StackName": "amplify-trainerportal-dev-90853-authtrainerportal0a4ecb86-1DZBYAP6LDL7F",
            "Parameters": [
                {
                    "ParameterKey": "authRoleArn",
                    "ParameterValue": "arn:aws:iam::795048271754:role/amplify-trainerportal-dev-90853-authRole"
                },
				...
```
				
I save the output in `test/cloudformation.json`. If a call the cfn api with "describeStacks" I would get a similar response back. 

The test gives a mocked response, which is the content of the json file. The response is mocked.

## Building "counter.go" skeleton

Now I create a `counter.go` file which at the beginning just defines the interface and a `Count` function which returns zero. This will give a failed test.

```go
type CounterInterface interface {
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)

}
```

```go

func Count(client CounterInterface) (int){
	
	return 0
}
```

This is the first step of the test-driven approach.

1) write a test
2) let test fail
3) write code until test passes

## Test-Driven-Development

### 1 - Write a test

I want to get a little help from friends. So I use a mocking framework `moq` which generates a helper class:

```go
//go:generate moq -out counter_moq_test.go . CounterInterface
```

With this remark the command `go generate` generates a file `counter_moq_test.go` which uses the interface definition to generate some helper classes. In the `counter_moq_test.go` there is also a generated documentation how to use the test!

With this help from the helper moq friend, I create the test file `counter.test.go`. Start the name of the test function with capital letters (remember?):

```go
func TestCountStacks(t *testing.T) {
	expectedValues := 2;
```

The test function defines the mock "DescribeStacks", which takes the saved json and return it:

```go
var cloudformationOutput cloudformation.DescribeStacksOutput
// Read json file
data, err := ioutil.ReadFile("test/cloudformation.json")
...
json.Unmarshal(data, &cloudformationOutput);
return &cloudformationOutput,nil;
```

Here the **strongly** typed nature comes into play: There is a structure for the input of the `DescribeStacks` and also for the response output. These lines take the response event `cloudformation.json` and transforms it into a structure aka "Unmarshalling".

This is the main "trick": I can now create different json files for corner cases which I want to test. The real counter code "thinks" the cloudformation API itself has send the response.

Now the test function calls the - to be implemented - Counter:

```go
computedValue := letsbuild13.Count(mockedCounterInterface)
```

To be able to do that the test imports the "letsbuild13" package.

Because the call passes the mocked client as an client, the Count function will call the mock client and will get the response defined in the `test/cloudformation.json` file.

Now comes the test assertion:

```go
assert.Equal(t,expectedValues, computedValue)
```

The Count function *should* return "2", because there are two stacks defined in the `test/cloudformation.json` .


### 2 - Let test FAIL

Now I start `go test` and get a fail, because at the moment the Count function returns 0.

```bash
go test
--- FAIL: TestCountStacks (0.00s)
    counter_test.go:38:
        	Error Trace:	counter_test.go:38
        	Error:      	Not equal:
        	            	expected: 2
        	            	actual  : 0
        	Test:       	TestCountStacks
FAIL
exit status 1
FAIL	letsbuild13	0.178s
```

### 3 - Write code until test passes

Then I write the code of the "Count" function until the test passes.

```go
input := &cloudformation.DescribeStacksInput{}
resp, _ := client.DescribeStacks(context.TODO(), input)
count := len(resp.Stacks)
return count
```

```bash
go test
PASS
ok  	letsbuild13	0.133s
```

![Celebrate](/img/2020/11/ray-hennessy-gdTxVSAE5sk-unsplash.jpg)

Because the response contains the `Stacks` structure as an array, `Count` just have to count the number of items (Stacks) in the array, to know how many CloudFormation stacks are deployed in the account.

If the test passes, that means the business functionality works. Now the main function is simple.

## Main 

At first main needs an `aws.config` class, which is used to initialize the real client.

With that `config` the cloudformation client is created.

```go
	cfg, err := config.LoadDefaultConfig(config.WithRegion("eu-central-1"))
    if err != nil {
        panic("unable to load SDK config, " + err.Error())
	}
	
	client := cloudformation.NewFromConfig(cfg);

	count := letsbuild13.Count(client);

	fmt.Println("Counting CloudFormation Stacks: ",count)
```


```bash
go run main/main.go
Counting CloudFormation Stacks:  8
```

## One more thing

Another advantage of golang is that you can compile static linked binaries for different operating systems.

Developing on mac, this command creates a mac binary:

```bash
go build -ldflags="-s -w" -o dist/cfn-count main/
```

After the build you may use `cfn-count` as standalone program:


```bash
./dist/cfn-count
Counting CloudFormation Stacks:  8
```

No problems with runtimes or a wrong python version etc.

Using go as lambda functions, you have to compile for linux. Thats easy:

```bash
env GOOS=linux go build -ldflags="-s -w" -o dist/linux/cfn-count main/main.go
```

You now can run the `cfn-count` on linux machines.

Same with windows, just change:

```bash
GOOS=windows GOARCH=amd64
```

And that`s it!

Thanks for reading, please comment on twitter. 
And visit/subscribe our twitch channel: [twitch](https://www.twitch.tv/tecracer).

Stay healthy in the cloud and on earth!

## Code

The code for this post is available here:

[https://github.com/megaproaktiv/aws-community-projects/tree/main/stackcount](https://github.com/megaproaktiv/aws-community-projects/tree/main/stackcount)

## Thanks

<span>Photo by <a href="https://unsplash.com/@rayhennessy?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText">Ray Hennessy</a> on <a href="https://unsplash.com/s/photos/fireworks?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText">Unsplash</a></span>