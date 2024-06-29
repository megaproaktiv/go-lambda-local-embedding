---
title: "Do you do Lambda Spaghetti?"
author: "Gernot Glawe"
date: 2022-02-25
draft: false
image: "img/2022/02/lambda-spaghetti.png"
thumbnail: "img/2022/02/lambda-spaghetti.png"
toc: true
keywords:
    - coding
    - lambda
    - go
    - python
    - javascript
tags:
    - level-200
    - lambda

categories: [aws]

---

Last week in the AWS slack developer channel once again, somebody was asking: "How can I run a Lambda locally?". Well, that is a valid question, but there is a chance that you only think you need a local Lambda emulator because you do Lambda Spaghetti! Spaghetti code is a phrase for unstructured and difficult-to-maintain source code. I show you an easier way to test Lambdas locally and have some arguments that a local Lambda runtime should only be the very last resort. On top, you get examples in Pasta-Python, Gemelli-Go and Tortellini-Typescript.

<!--more-->


## The problem of spaghetti Lambda

Overcooking pasta gives you sticky and clumpy pasta. Running Lambda functions as a whole gives you a monolith.

If you write Lambda code that is longer than approximately three logical steps, it should not be coded in a single function! This is monolithic programming on the code level, and the parts should be separated into functions within the same Lambda. I show you why:

![Mini Monolith](/img/2022/02/sl-test1.png)
Have a look at the diagram of this mini-monolith.


You can only test all functions together, so they are tightly coupled. If you do testing of the whole block (the lambda handler) and get an error, you do *not know* in which part of the application the error occurs. In Addition, you need the emulation of the Lambda compute resource. This resource could be emulated locally with [**lambci**](https://github.com/lambci) or [**sam-cli**](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-local-invoke.html). But this often involves running a docker container locally, which introduces additional time and consume additional local computing resources.


## The first part of the solution is to run Lambda functions without the Lambda resources.

![Mini Monolith without lambda](/img/2022/02/sl-test2.png)

To have a fast and easy test, you directly invoke the code without any Lambda resources/environment.

### Lambdaless Python

We have this small Lambda python function:

```py
import json

def lambda_handler(event, context):
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
```

To call it and get the return values, you just need a tiny script "call.py"

```py
import index

ret = index.lambda_handler({}, {})
print(ret)
```

Run it and you get:

```bash
python call.py
{'statusCode': 200, 'body': '"Hello from Lambda!"'}
```

That's all, no containers needed!

That is the same approach as the "serverless framework" does with [serverless invoke local](https://www.serverless.com/framework/docs/providers/aws/cli-reference/invoke-local).

### Lambdaless JavaScript

```js
exports.handler = async (event) => {
    // TODO implement
    const response = {
        statusCode: 200,
        body: JSON.stringify(event.key1),
    };
    
    console.log(extract.extract(event))
    
    return response;
};
```

This is the Lambda handler.

```js
'use strict';

const fs = require('fs');
const lambda = require('./lambda')

let rawdata = fs.readFileSync('event.json');
let event= JSON.parse(rawdata);
lambda.handler(event);
```

You can call the handler with this piece of code. If you have the incoming event stored as a file locally, you can give it as a parameter directly. More on this later.

## The second part of the solution is to test methods from the functions independently


![Mini Monolith without lambda](/img/2022/02/sl-test3.png)

Suppose we have a Lambda function that calculates the [Body Mass Index](https://en.wikipedia.org/wiki/Body_mass_index). With a pasta Lambda you would write the calculation directly into your handler.

The better way is to write a small module, which you can test independently:

```js
function bmi(age, weight, height) {
  height_m = height/100;
  bmi = weight / (height_m * height_m);
  bmi = Math.round(bmi);
  return bmi;
}
module.exports = bmi;
```

`bmi.js`

In your lambda handler you call this method:

```js
const bmi=require('./bmi');

// Handler
exports.handler = async function(event, context) {

  age = event.age;
  weight= event.weight;
  height= event.height;
  return bmi(age, weight, height);
}
```

`index.js`

Now you can call and test this bmi function independently, or even better write a unit test:

```js
const bmi = require('./bmi');

test("Bmi Übergewicht", () => {
  expect(bmi(42,113,188)).toBe(32);
});
```

This example is done with the [jest](https://jestjs.io/) testing framework.

### Python Testing

The  same can be achieved with python, e.g. with [pytest](https://docs.pytest.org/en/7.0.x/)

### GO Testing

For a detailed intro see [go-on-aws](https://www.go-on-aws.com/testing-go/)

## How to get real test data?

To test a method, you have to give some test parameters to the function. The Lambda functions on AWS will be invoked with a JSON event. How to get this data?

### Get the json event

The trick is to write a dump function at first and save the data like this:

```js
exports.handler = async (event) => {
    console.log(JSON.stringify(event))
    const response = {
        statusCode: 200,
        body: JSON.stringify('Hello from Lambda '),
    };
    return response;
};
```

### Get testdata from AWS API calls

If it's an AWS call, you can just call the API with the AWS CLI and save the JSON as a file.
E.g. if you want to have the describe instances API call:

```bash
aws ec2 describe-instances >event.json
```

## Testlevels

Your test concept should ensure that you test specific functionality without side effects. And that you test under a near-real production environment.

Running Lambda functions as a whole locally on your workstation does none of these.

### Test specific functionality

This is just to test the "does it work".

- Test single functions
- Start with coded tests, consider using a test framework after some time

### Test Lambda as a whole

This is to test "does everything work together" and "does it work in the real environment".

- Invoke Lamvda on AWS
- Test the Lambda IAM rights also
- Have production latency and timing
- Have limits for accessing disc (`/tmp` space)
- Have limits for CPU cores
- Have limits for memory


## Summary

To begin non-monolith or non-spaghetti programming, you don't have to start with unit tests and a test framework right away if you want to avoid the learning curve. You just refactor your methods from the lambda handler and call them one by one. This way, you will write better code!

I hope you got some appetizer for doing unit testing instead of running your whole lambda locally! So follow me on Twitter and visit my GO on AWS site if you haven't done it already.


## Feedback & discussion

For discussion please contact me on twitter @megaproaktiv

## Learn more GO

Want to know more about using GOLANG on AWS? - Learn GO on AWS: [here](https://www.go-on-aws.com/)

