---
title: "Authenticate local react app with static AWS credentials"
author: "Gernot Glawe"
date: 2024-02-11
draft: false
image: "kyle-glenn-dGk-qYBk4OA-unsplash.jpg"
thumbnail: "kyle-glenn-dGk-qYBk4OA-unsplash.jpg"
toc: true
keywords:
    - bedrock
    - go
    - python
    - kendra
tags:
    - level-300
    - node-sdk
    - authentication
    - react
categories: [development]
---


# Local browser app development, SDKs and credential APPS:

## Problem:

You want to develop a local browser app with the AWS SDK and you want to use your local AWS credentials. Although your current credentials are valid, the SDK does not accept them.

## Solution:

Use cognito or use framework specific solutions to provide `ENV` variables to the SDK.

## Prerequisites:

Checkin on `bash`/ `cli` if your credentials are valid:

```bash
aws sts get-caller-identity
```

Output:

```json
{
    "UserId": "AROA******M3AGZ4:soldier",
    "Account": "939*****8838",
    "Arn": "arn:aws:sts::93***838:assumed-role/umbreallaCorpSuperAdminC/soldier
}
```

## Part 1: Python/streamlit "I am gonna take the local aws creds"
## Part 2: Node SDK

AWS Authentication in the node sdk is done as a channel, several authentication types are checked.

Most of them are not available in the browser or native apps.

See [credential providers](https://docs.aws.amazon.com/AWSJavaScriptSDK/v3/latest/Package/-aws-sdk-credential-providers/#fromini)

```txt
fromEnv()
Not available in browser & native apps.
---
fromIni()
May use @aws-sdk/client-sso or @aws-sdk/client-sts depending on how the file is configured.
Not available in browsers & native apps.
---...

```


## Workaround (just for local development) with REACT:
Shell Script

```bash
echo REACT_APP_KEY_NAME=$AWS_ACCESS_KEY_ID >.env
echo REACT_APP_SECRET_NAME=$AWS_SECRET_ACCESS_KEY >>.env
echo REACT_APP_TOKEN_NAME=$AWS_SESSION_TOKEN >>.env
cp .env .env.local
cp .env .env.development.local
```

Then in node/browser app:

```javascript
export const lambdaClient = new LambdaClient({
    region: "eu-central-1",
    credentials:{
        accessKeyId:process.env.REACT_APP_KEY_NAME,
        secretAccessKey:process.env.REACT_APP_SECRET_NAME,
        sessionToken: process.env.REACT_APP_TOKEN_NAME
    }
 });
``

** Note: put the `.env` file in your `.gitignore` file.**
