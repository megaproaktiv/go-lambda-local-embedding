---
title: "The cuckoo egg testing lambda"
author: "Gernot Glawe"
date: 2023-06-20
draft: false
image: "img/2023/06/cuckoo.jpeg"
thumbnail: "img/2023/06/cuckoo.jpeg"
toc: true
keywords:
    - lambda
    - serverless
tags:
    - level-200
categories: [aws]

---

Oh, there is an error in my Lambda function. But - what is the event JSON input which caused the error? Oh, I forgot to log the event in my Lambda code. Damned! It would be great to swap the code with a "just dump the event code" and slip it like a cuckoo egg. Afterwards, get the event and restore the old Lambda!
 
<!--more-->

No problem!

## Swap
Only some people know that you can also download the function code of a Lambda function. This can be done via the `Code.Location` property in the `get-function` API call. This property contains a presigned URL to download the function code.

That means we could save the old function code and swap it with the test "cuckoo egg" function code. Then we store the event and restore the function afterwards.

You could also use this pattern to replace Lambda functions in a distributed system with mock Functions, which respond with a fixed value.

![swap](/img/2023/06/swap.png)

## Walkthrough

These are the steps:

1) Save old Lambda, download code

```bash
aws lambda get-function
        --function-name "{{.NAME}}"
        --query 'Code.Location'
        --output text >download.txt
wget -i download.txt -O {{.CODE}}
```

You must also store the Function Name, the Runtime and the handler. The language and the handler will often be the same as in your other Lambda Functions.


2) Deploy egg

```bash
  aws lambda update-function-configuration --runtime go1.x --function-name {{.NAME}} --handler main
  aws lambda update-function-code --function-name  {{.NAME}} --zip-file fileb://./dist/main.zip
```

The function for the event is of course written in GO. Should be easy to port to other languages.

3) Create event, get data from CloudWatch logs

4) Restore old code:

```bash
aws lambda update-function-configuration --runtime {{.RUNTIME}} --function-name {{.NAME}} --handler {{.HANDLER}}
aws lambda update-function-code --function-name  {{.NAME}} --zip-file fileb://{{.CODE}}
```

All steps are coded in the `Taskfile.yml` file.
See the [Repo on github](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-dump-event)

## Usage with `taskfile.dev`

See [Taskfile.yaml](https://github.com/megaproaktiv/aws-community-projects/blob/main/lambda-dump-event/Taskfile.yml)

A) Get code from [github](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-dump-event)

B) Update the `Taskfile.yml` with your values:

```yaml
NAME: polly-notes-api-searchFunction
RUNTIME: python3.8
HANDLER: app.lambda_handler
CODE: code.zip
```

These are example values.

1) Save:  `task save`
2) Deploy: `task deploy`
3) Invoke Lambda with event
4) Restore: `task restore`

## Save

```bash
task save
task: [save] aws lambda get-function --function-name "polly-notes-api-searchFunction" --query 'Code.Location' --output text >download.txt
task: [save] wget -i download.txt -O code.zip
--2023-06-19 18:07:27--  https://awslambda-eu-cent-1-tasks.s3.eu-central-1.amazonaws.com/snapshots/139008737997/polly-notes-api-searchFunction-***
&X-Amz-Credential=ASIAZ***%2Feu-central-1%2Fs3%2Faws4_request
&X-Amz-Signature=00***7
Auflösen des Hostnamens awslambda-eu-cent-1-tasks.s3.eu-central-1.amazonaws.com (awslambda-eu-cent-1-tasks.s3.eu-central-1.amazonaws.com)… 52.219.171.254
Verbindungsaufbau zu awslambda-eu-cent-1-tasks.s3.eu-central-1.amazonaws.com (awslambda-eu-cent-1-tasks.s3.eu-central-1.amazonaws.com)|52.219.171.254|:443 … verbunden.
HTTP-Anforderung gesendet, auf Antwort wird gewartet … 200 OK
Länge: 12042877 (11M) [application/zip]
Wird in »code.zip« gespeichert.

code.zip                                    100%[========================================================================================>]  11,48M  23,9MB/s    in 0,5s

2023-06-19 18:07:28 (23,9 MB/s) - »code.zip« gespeichert [12042877/12042877]

BEENDET --2023-06-19 18:07:28--
Verstrichene Zeit: 1,0s
Geholt: 1 Dateien, 11M in 0,5s (23,9 MB/s)
```



## Deploy

```bash
task deploy
updating: main (deflated 58%)
task: [deploy] aws lambda update-function-configuration --runtime go1.x --function-name polly-notes-api-searchFunction --handler main
{
    "FunctionName": "polly-notes-api-searchFunction",
    ...
    "Runtime": "go1.x",
    ...
}
task: [deploy] aws lambda update-function-code --function-name  polly-notes-api-searchFunction --zip-file fileb://./dist/main.zip
{
    "FunctionName": "polly-notes-api-searchFunction",
    "Runtime": "go1.x",
    "Handler": "main",
    "CodeSize": 2691074,
    ...
    "PackageType": "Zip",
    ...
}
```

## Restore

```bash
task restore
task: [restore] aws lambda update-function-configuration --runtime python3.8 --function-name polly-notes-api-searchFunction --handler app.lambda_handler
{
    "FunctionName": "polly-notes-api-searchFunction",
    ...
    "Runtime": "python3.8",
    ...
    "Handler": "app.lambda_handler",
    ...
}
task: [restore] aws lambda update-function-code --function-name  polly-notes-api-searchFunction-GdFQpjWz1XKZ --zip-file fileb://code.zip
{
    "FunctionName": "polly-notes-api-searchFunction",
    ...
    "Runtime": "python3.8",
    ...
    "Handler": "app.lambda_handler",
    "CodeSize": 12042877,
    ...
}
```

## Conclusion

With this little trick, you can inject test code into a running Lambda based microservice system and restore the old environment afterwards.


Happy building!


If you need consulting for your serverless project, don't hesitate to get in touch with the sponsor of this blog, [tecRacer](https://www.tecracer.com/kontakt/).

For more AWS development stuff, follow me on dev https://dev.to/megaproaktiv.
Want to learn GO on AWS? [GO here](https://www.go-on-aws.com/)

## See also

- [Source code](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-dump-event)
