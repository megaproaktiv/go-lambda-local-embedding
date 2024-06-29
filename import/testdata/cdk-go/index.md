---
title: "Easy going - programming AWS Resources with the CDK in GO"
author: "Gernot Glawe"
date: 2021-04-07
draft: false
image: "img/2021/04/balkouras-nicos-8tadBT_8Gag-unsplash.jpg"
thumbnail: "img/2021/04/balkouras-nicos-8tadBT_8Gag-unsplash.jpg"
toc: true
keywords:
    - go
    - vpc
    - cdk
    - level-200
tags:
    - level-200
    - go
    - vpc
categories: [aws]
---

# CDK GO is in preview, here are the pros and cons and a first VPC example

TL;DR The GO module system is IMHO neater than node.JS or Python. But you have to get used to the strongly typed language GO.

<!--more-->

## Init App

In CDK you generate the app skeleton with a single command. The generation of the GO app is really fast ( under one second):

```bash
mkdir go-vpc && cd go-vpc
cdk init app --language=go
```

## Files

Name | Description
---|---
cdk.json  | the usual cdk configuration
go-vpc.go  | this is the single app file
go-vpc_test.go  | testing, called with `go test`
go.mod | this is the go package configuration
README.md | all you need to know




## Import constructs for VPC

You change the "sns" import to ec2 in the `import` section of `go-vpc.go` (line 5):

```go
  3 import (
  4         "github.com/aws/aws-cdk-go/awscdk"
  5         "github.com/aws/aws-cdk-go/awscdk/awsec2"
  6         "github.com/aws/constructs-go/constructs/v3"
  7         "github.com/aws/jsii-runtime-go"
  8 )
  ```

You may download each import with:

`go get github.com/aws/aws-cdk-go/awscdk/awsec2`

Or you get all imports automatically with 

`go mod tidy`

Or you just let the CDK handle that!

The app configuration in `cdk.json` is:

`"app": "go mod download && go run go-vpc.go",`

1) `go mod download` - this handles all imports
2) `go run go-vpc.go` - this compiles and runs the app

So, if you just start a `cdk diff` or `cdk ls`, all modules will be downloaded. 
More information about the GO module system gives: `go mod help`.

## Define VPC

Do I have to mention where? 

Just add some code after `	// The code that defines your stack goes here `


```go
awsec2.NewVpc(stack, jsii.String("MyVpc"),
        &awsec2.VpcProps{
                Cidr: jsii.String("10.0.0.0/16"),
        },
)
``` 



## Call diff

```bash
cdk diff
```

## Call deploy

```bash
cdk deploy
```

That was easy :) .

No npm or pip actions, just define import settings and go (pun intended)

## Differences to TypeScript

### TypeScript is strongly typed

In TS the VPC definition looks like this:

![](/img/2021/04/2021-04-07_08-28-08.png)

1) Strings
  -   Most strings are - strings
2) properties
  - TS knows that `{ cidr: "10.0.0.0./16" } ` is a valid VPC property. But you could add properties that are not valid.
  - the property values are just - strings

### GO is *really* strongly typed

In GO the type system much more rigid:

![](/img/2021/04/2021-04-07_08-20-35.png)

1) String pointer
  Just like in the AWS GO SDK with the "aws.String" helper, jsii helps you in converting the "MyVpc" string to a string pointer

The definition of the NewPVC is:

```go
func awsec2.NewVpc(scope constructs.Construct, id *string, props *awsec2.VpcProps) awsec2.Vpc
```

So `id` is a "*string", which means an pointer to a string. Usually, you would have to do:

```go
myid := "MyVpc"
...
awsec2.NewVpc(stack, &myid...)
```

2) You have to define explicitly define that vpc props are VpcProps.
This may seem overly complicated, but it enforces the right properties and makes some IDE help easier to implement.

![](/img/2021/04/fill.png)

### TypeScript is an addition to JavaScript

And so we have some more files just for the TypeScript Compiler (tsconfig.json) and the external testing system (jest).

![](/img/2021/04/2021-04-07_08-28-58.png)

### GO has all included

With GO you have fewer files:

![](/img/2021/04/2021-04-07_08-28-46.png)

## Fast >> slow >> Virusscan

I was a little bit disappointed that go cdk is not faster, but much slower than typescript. But as it turned out, not GO is to blame, but my Virus Scan module. So i checked the timing on AWS cloudshell (tm). With cloudshell you just have to install go 1.16, because yum still has go 1.15.

Environment | Language | Command | Time
---|---|---|---
Mac with Sophos Virus | go | cdk diff | 20 sec
Mac with some Scan disabled | go | cdk diff | 10 sec
Mac with Sophos Virus | ts | cdk diff | 7 sec
Mac with some Scan disabled | ts | cdk diff 3 sec
cloudshell | ts | init app | 33 s
cloudshell | ts | ` npm i @aws-cdk/aws-ec2` | 10 sec
cloudshell | ts | ` cdk diff` | 7 sec
cloudshell | go | init app | 1 s
cloudshell | go | `cdk diff`  first time | 30 s
cloudshell | go | `cdk diff`  second time | 7 s

## Ready to go

Again the CDK team and contributors around the world have done a tremendous job!

With CDK go the package handling is much more smoother and if you get used to the really strong type system, maybe you should give it a try!

## Thanks to

Photo by <a href="https://unsplash.com/@ba1kouras?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Balkouras Nicos</a> on <a href="https://unsplash.com/s/photos/walk?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Unsplash</a>
  
