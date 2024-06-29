---
title: "Putting the database to sleep using Lambda - a Python developer's first contact with Golang"
author: "Maurice Borgmeier"
date: 2022-05-11
toc: false
draft: false
image: "img/2022/05/chinmay-bhattar-fd9mIBluHkA-unsplash.jpg"
thumbnail: "img/2022/05/chinmay-bhattar-fd9mIBluHkA-unsplash.jpg"
categories: ["aws"]
tags: ["level-200", "lambda", "rds", "go"]
summary: |
    In this blog, I take you along on my journey to build my first Golang-based Lambda function.
    Inspired by surprise on my RDS bill, I built a Lambda function in Go to periodically stop running databases with a specific tag.
    Come, learn and debug with me!

---

Today's blog is inspired by my AWS bill, my research list, one of Corey Quinns' [recent blog posts](https://www.lastweekinaws.com/blog/shitposting-as-a-learning-style/), and a talk [by Uncle Bob I watched](https://www.youtube.com/watch?v=2dKZ-dWaCiU). While working with a customer, I set up a couple of RDS databases for performance tests. I shut them down after I was done with the intention of restarting them a few days later for additional tests.

Projects being projects, things got delayed a little bit, and after a while, I noticed that my monthly AWS bill was higher than usual. Sure enough, the database instances were running. I had forgotten that AWS [will restart stopped RDS instances after 7 days](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html#:~:text=Important,required%20maintenance%20updates.) to apply updates (or to mess with you). After shutting them down again, I decided to write a Lambda function to stop the DBs if they were started again. Since I was interested in learning Golang, I decided to use that for a change and take you with me on the journey.

I'm usually at home in the Python world, and here, I would have written something like this:

```python
# Pseudocode
def lambda_handler(event, handler):
    rds_instances = get_rds_instances()
    for instance in rds_instances:
        if instanceShouldBeStopped(instance):
            stop_rds_instance(instance)
```

My goal was to replicate that in Go. I first needed to install the language on my Mac to get going. Note that I'm using a Mac with an ARM-based processor. This will become relevant later. Installing Go was a breeze using brew.

```terminal
$ brew install golang
```

Having installed Go, I tried to write Hello World with code from [go by examples](https://gobyexample.com/hello-world):

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
```

Writing the code worked, and running it was easy after I figured out that the `.go` suffix in the command was essential. If you omit it (like I did at first), you'll see a `package main is not in GOROOT` error, which isn't all too helpful.

```terminal
$ go run main.go
Hello World
```

We can also compile the code to a binary and run it, but at this point, that's just extra steps I'm not interested in. I want to build stuff. If we wanted to do that, this is how that works:

```terminal
$ go build main.go
$ ./main
Hello World
```

First, I want to play around with the AWS SDK for Go and find a way to list my currently running RDS instances and their tags. Apparently, we *should* be able to just download it using the following command that's documented [on the SDK's Github page](https://github.com/aws/aws-sdk-go). Well, that didn't work. It wants a `go.mod` file, which is apparently used to track dependencies and their versions in a Go module.

```terminal
$ go get github.com/aws/aws-sdk-go
go: go.mod file not found in current directory or any parent directory.
        'go get' is no longer supported outside a module.
        To build and install a command, use 'go install' with a version,
        like 'go install example.com/cmd@latest'
        For more information, see https://golang.org/doc/go-get-install-deprecation
        or run 'go help get' or 'go help install'.
```

In the official [getting started docs](https://go.dev/doc/tutorial/getting-started), I found a command to do that: `go mod init example/hello`. Substituting the project name for my own allowed me to create a package and subsequently install the SDK:

```terminal
$ go mod init mauricebrg/rds-sleep
go: creating new go.mod: module mauricebrg/rds-sleep
go: to add module requirements and sums:
        go mod tidy
$ go get github.com/aws/aws-sdk-go
go: downloading github.com/aws/aws-sdk-go v1.44.6
go: downloading github.com/jmespath/go-jmespath v0.4.0
go: added github.com/aws/aws-sdk-go v1.44.6
go: added github.com/jmespath/go-jmespath v0.4.0
```

Afterward, there are two more files in my directory. The `go.mod` seems to track the installed dependencies, and the `go.sum` appears to have checksums for each installed dependency. My next goal is to instantiate an RDS service client and list the database instances. After some playing around, I managed to do it:

```go
package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func main() {

	awsRegion := "eu-central-1"

	// Apparently, we need to create a session first
	// Must makes things crash if something goes wrong
	session := session.Must(session.NewSession(&aws.Config{Region: &awsRegion}))

	// We then use this session to get an rds client
	rdsClient := rds.New(session)

	response, err := rdsClient.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		// Do some error handling
		fmt.Println(err)
		os.Exit(1)
	}

	// The range does something like enumerate() in python
	for _, dbInstance := range response.DBInstances {

		instanceId := *dbInstance.DBInstanceIdentifier
		instanceStatus := *dbInstance.DBInstanceStatus

		fmt.Println(instanceId, "is in status", instanceStatus)

	}

}
```

I learned a few things during this process:

- [Pointers](https://go.dev/tour/moretypes/1) are fun (not really).
- Go is particular about variable names and recommends you use `camelCase` aka. `mixedCase` for variable names.
- Golang doesn't have exceptions. The common way to handle errors is by returning them as the second return value. You need to check if the call was successful.

In order to use this in my future Lambda function, I encapsulated this API call in a function:

```go
func listDBInstances() ([]*rds.DBInstance, error) {
	awsRegion := "eu-central-1"

	// Apparently we need to create a session first
	// Must makes things crash if something goes wrong
	session := session.Must(session.NewSession(&aws.Config{Region: &awsRegion}))

	// We then use this session to get an rds client
	rdsClient := rds.New(session)

	response, err := rdsClient.DescribeDBInstances(&rds.DescribeDBInstancesInput{})

	return response.DBInstances, err
}
```

In the same fashion, I also implemented `stopDBInstance`, which wraps the respective API call, and `dbInstanceShouldBeStopped`, which returns true for running instances also a tag that tells us to stop them. This allows me to implement my `putDBInstancesToSleep` function as follows:

```go
func putDBInstancesToSleep() error {

	dbInstances, err := listDBInstances()
	if err != nil {
		return err
	}

	for _, dbInstance := range dbInstances {
		if dbInstanceShouldBeStopped(dbInstance) {
			err := stopDBInstance(dbInstance.DBInstanceIdentifier)
			if err != nil {
				return err
			}
		}
	}

	return nil

}
```

Aesthetically the code doesn't look as pleasing to me as the pseudocode in Python - most likely because of all the error checking here. In Python, exceptions cause the function to crash unless they're caught and handled, which is fine for my use case here. That's probably not a good practice, though. I guess this is one of those things that takes some getting used to.

Now that I've got code that works locally, it's time to get it into a Lambda function. Looking into the [documentation](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html), it seems that I'll first need to install another package.

```terminal
$ go get github.com/aws/aws-lambda-go/lambda
go: downloading github.com/aws/aws-lambda-go v1.31.1
go: added github.com/aws/aws-lambda-go v1.31.1
```

From Python, I'm used to having a simple `lambda_handler` function that receives the event as a dictionary and the context object (which I usually don't care about). Golang is a bit more specific here and wants me to define the event's structure that will be handed to the handler. Fortunately, there are pre-built structs available for common event sources. Since I want this to be invoked on a schedule via CloudWatch Events / EventBridge, I choose the adequate struct for the event. Also, we can choose to return nothing, an error, or response **and** an error. Since this code will only be triggered from CloudWatch events, we don't need a response. That leads to the following implementation.

```go
import (
	//...
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

//...

func HandleLambdaEvent(event events.CloudWatchEvent) error {
	return putDBInstancesToSleep()
}

func main() {

	lambda.Start(HandleLambdaEvent)

}
```

Time to bundle everything up and create our deployment package [according to the documentation](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html#golang-package-mac-linux)!

```terminal
$ # First, we build the go binary for Linux
$ GOOS=linux go build main.go
$ # Time to zip the binary
$ zip function.zip main
```

Next, I create a new Lambda function in the AWS Console and upload the archive.

![Create Lambda View](/img/2022/05/go_lambda1_0.png)

Then I uploaded the ZIP archive we created earlier.

![Upload ZIP Archive](/img/2022/05/go_lambda1_1.png)

Next, I define this test event:

```json
{
  "id": "cdc73f9d-aea9-11e3-9d5a-835b769c0d9c",
  "detail-type": "Scheduled Event",
  "source": "aws.events",
  "account": "123456789012",
  "time": "1970-01-01T00:00:00Z",
  "region": "us-east-1",
  "resources": [
    "arn:aws:events:us-east-1:123456789012:rule/ExampleRule"
  ],
  "detail": {}
}
```

Running this Test event returns an error:

![fork/exec no such file or directory PathError](/img/2022/05/go_lambda1_3.png)

Apparently, it expects the handler to be called `hello` when you create the function through the GUI. No problem, we can change that to `main` in the Runtime Settings.

![Lambda Runtime Settings](/img/2022/05/go_lambda1_2.png)

After doing that, running the test event yields a different error:

```json
{
  "errorMessage": "fork/exec /var/task/main: exec format error",
  "errorType": "PathError"
}
```

Strange. Fortunately, [stackoverflow](https://stackoverflow.com/questions/50700979/exec-format-error-when-running-aws-golang-lambda) has an answer. I forgot I was running this on an M1-based Mac, so it was compiled for ARM. The Go runtime doesn't (yet) support ARM-based Lambdas, so I had to recompile the Lambda and update it.

```terminal
$ # First, we build the go binary for Linux
$ # This time, with the correct CPU architecture
$ GOARCH=amd64 GOOS=linux go build main.go
$ # Time to zip the binary
$ zip function.zip main
```

After updating the function code, I could finally run the Lambda function using my test event. Now I also learned how to compile binaries for different CPU instruction sets.
![Lambda Success](/img/2022/05/go_lambda1_4.png)

All that's left is to create a trigger to run this every day at 7 pm.

![Lambda Create CloudWatch trigger](/img/2022/05/go_lambda1_5.png)

Granted, this is not infrastructure as code or automated, but it's a start. More things I want to add include logging and unit testing. It works, but I'd like to be more confident that it will continue to do so in the future.

I built a Lambda function in Go that shuts down RDS databases with a predefined Tag every day at 7 pm. This is nothing spectacular, but I learned a lot about Golang and may tackle optimizing the setup in a future blog post. If you're interested, you can find the code I've shown you [here](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/first-go-lambda).

I hope you learned something as well, and I'm looking forward to your feedback. Feel free to reach out to me via the channels mentioned in my bio.

&mdash; Maurice

(Photo by [Chinmay Bhattar](https://unsplash.com/@geekgunda?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/golang?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText))