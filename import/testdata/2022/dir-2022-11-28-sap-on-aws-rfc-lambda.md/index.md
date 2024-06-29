---
title: "Call SAP RFC Function Modules from AWS Lambda using the NW RFC SDK and Node.js"
author: "Fabian Brakowski"
date: 2022-11-28
toc: true
draft: false
image: "img/2019/09/fleur-dQf7RZhMOJU-unsplash.jpg"
thumbnail: "img/2019/09/fleur-dQf7RZhMOJU-unsplash.jpg"
categories: ["aws", "sap"]
tags: 
    - sap
    - level-300
    - lambda
    - nwrfcsdk
    - sap rfc
keywords:
    - sap
    - lambda
    - nwrfcsdk
    - sap rfc

    
    
---

SAP provides the SAP NetWeaver RFC SDK to allow external systems to call SAP Systems directly via SAP RFC. I will walk you through the steps it took me to deploy a function on AWS Lambda that uses the NW RFC SDK. Once you mastered those, you could use it to call any remote-enabled Function Module from AWS Lambda.

<!--more-->

## NetWeaver RFC SDK

SAP RFC (Remote Function Call) is the primary communication protocol for inter-system communication in SAP ABAP Systems. When programming in ABAP itself, it is very easy to use as it is natively integrated. However, as it is a proprietary protocol, there is no native integration for other platforms. This is where the NetWeaver RFC SDK comes in. It is a library, written in C++, that provides programming interfaces for using RFC outside of ABAP. For languages other than C++, they provide so-called client bindings like ``node-rfc`` for Node.js or ``PyRFC`` of Python. This allows native integration into other programming languages, while still calling the C++-library in the background.

Why do I want to use old-school RFC over other, more modern, communication protocols? Indeed, in the last few years SAP itself has integrated many open communication protocols into its platform, e.g. SOAP and OData. And they, as well as their customers have used those protocols to build well-functioning integrated systems. However, due to its long history, much more RFC Function Modules are available by default than Web Service Methods. This is especially true for systems administration purposes. While I could write the needed ABAP code to port those functions myself, I prefer to stick with the standard implementations provided by SAP. Additionally, RFC Function Modules usually work out of the box. Little configuration is needed on the server side, unlike with the OData Gateway which requires many additional setup steps.

## Environment requirements

SAP provides the NW RFC SDK as C++-libraries. For Node.js (or other runtime environments) to use it, those libraries need to first be installed and then made available for the runtime environment.
The zip file has to be downloaded from SAP Marketplace for licensing reasons. On linux, it is recommended to deploy them into ``usr/local/sap/nwrfcsdk`` and then ensure that the dynamic link loader can find it. For this, either edit the environment variable ``LD_LIBRARY_PATH`` or create a file ``/etc/ld.so.conf.d/nwrfcsdk.conf``. In both cases simply specify the library directory, e.g. ``usr/local/sap/nwrfcsdk/lib``. For the second option, run the command ``ldconfig`` to load the new configuration. 


## Setting up the source code

After the system is prepared, set up the source code for a local test. Create a new directory, initialize it as a node project and then install ``node-rfc``.

```bash
mkdir sap-rfc-test
npm init
npm install node-rfc

```

## Local Testing

Before deploying any code on lambda, I prefer to make sure it works locally to rule out any programming errors. For a first test, I copied the [sample code](https://github.com/SAP/node-rfc#direct-client) and modified it. The function I am calling will output some details about my SAP system. The file name should be ``index.ts``.

```ts
const noderfc = require("node-rfc");
const client = new noderfc.Client({ dest: "P10" });

(async () => {
    try {
        await client.open();

        const result = await client.call("RFC_GET_SYSTEM_INFO", {
            DESTINATION: "NONE"
        });

        // check the result
        console.log(result);
    } catch (err) {
        // connection and invocation errors
        console.error(err);
    }
})();
```

For the code to work, we need to tell the code some connection details of system P10. For this purpose, place a file called ``sapnwrfc.ini`` into the same directory. Please note my comments about hard-coding system credentials that I put at the bottom of this blog post.

```ini
DEST=P10
USER=testuser
PASSWD=password
ASHOST=3.71.242.179
SYSNR=00
CLIENT=000
LANG=EN
```

If everything worked, you should see the results of the ``RFC_GET_SYSTEM_INFO`` RFC-Call.

```javascript
{
  CURRENT_RESOURCES: 6,
  DEST_COMMUNICATION_MESSAGE: '',
  DEST_SYSTEM_MESSAGE: '',
  DIALOG_USER_TYPE: '',
  MAXIMAL_RESOURCES: 8,
  RECOMMENDED_DELAY: 0,
  RFCSI_EXPORT: {
    RFCPROTO: '011',
    RFCCHARTYP: '4103',
    RFCINTTYP: 'LIT',
    RFCFLOTYP: 'IE3',
    RFCDEST: 'app-test2_P10_00',
    RFCHOST: 'app-test',
    RFCSYSID: 'P10',
    RFCDATABS: 'P10',
    RFCDBHOST: 'hdb-test2',
    RFCDBSYS: 'HDB',
    RFCSAPRL: '755',
    RFCMACH: '  390',
    RFCOPSYS: 'Linux',
    RFCTZONE: '     0',
    RFCDAYST: '',
    RFCIPADDR: '10.10.0.121',
    RFCKERNRL: '785',
    RFCHOST2: 'app-test2',
    RFCSI_RESV: '',
    RFCIPV6ADDR: '10.10.0.121'
  },
  RFC_LOGIN_COMPLETE: '',
  DESTINATION: 'NONE'
}
```

## Deploying the function on Lambda

### Modify the code for Lambda

Lambda requires the code to export a handler function that can be called by the Lambda runtime. To make the above code usable with Lambda, I modified it by wrapping the code inside such a handler function. As the function is asynchronous I am returning the Promise that is created by the function.

```ts
const noderfc = require("node-rfc");

export async function handler() {
  const client = new noderfc.Client({ dest: "P10" });

  return (async () => {
      try {
          await client.open();

          const result = await client.call("RFC_GET_SYSTEM_INFO", {
              DESTINATION: "NONE"
          });

          // check the result
          console.log(result);

          return result;
      } catch (err) {
          // connection and invocation errors
          console.error(err);
          return err;
      }
  })();
}
```

### Deploy as container

I decided to deploy the lambda function as a custom docker container. One reason is the need for the C++-library on the runtime. Another is the size of the ``node-rfc`` module itself.

Create a ``Dockerfile`` with the following contents:

```Dockerfile
FROM public.ecr.aws/lambda/nodejs:16
COPY index.ts package.json tsconfig.json sapnwrfc.ini ./ 
COPY nwrfcsdk /bin/local/sap/nwrfcsdk
ENV LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:/bin/local/sap/nwrfcsdk/lib"
RUN npm install -g typescript
RUN npm install
RUN tsc
CMD [ "index.handler" ]
```

To use it in Lambda, first create a new ECR repository and then push the locally build container to it. This documentation can help: [Link](https://docs.aws.amazon.com/AmazonECR/latest/userguide/getting-started-cli.html). 

I ran the following commands to log in to ECR, build the container and push it. Please ensure to change the account id to yours and authenticate with the AWS CLI before running the commands.


```Bash
aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.eu-central-1.amazonaws.com
docker build . -t 123456789012.dkr.ecr.eu-central-1.amazonaws.com/sap-rfc-test
docker push 123456789012.dkr.ecr.eu-central-1.amazonaws.com/sap-rfc-test
```

Afterwards, the container can be selected for deployment in Lambda.

![TODO](/img/2022/09/sap-on-aws-nwrfcsdk-deploy-lambda.png)

Or, if you are lazy, run:

```Bash
aws lambda update-function-code --function-name sap-rfc-test --image-uri 123456789012.dkr.ecr.eu-central-1.amazonaws.com/sap-rfc-test:latest
```

### Test results

After some problems with correctly configuring the dynamic link loader to find the C-libraries, the function now works as expected. The result is returned as JSON and can be easily be processed.

![TODO](/img/2022/09/sap-on-aws-nwrfcsdk-lambda-results.png)

## Considerations

When using the NetWeaver RFC SDK, please consider the following:

### SSM Parameter Store/ AWS Secrets Manager
 
All connection details in the above example are hard-coded into the ``sapnwrfc.ini`` file. In my case, this was uncritical, as the SAP system I used is a pure sandbox for playing around. Never (!!!) push actual credentials of any SAP System (incl. DEV and QAS) into a container repository or even worse a Git repository. With AWS, it is very easy to store such credentials, as well as connection details, in SSM Parameter Store and/or AWS Secrets Manager. During runtime, you can then easily pull those credentials and use them to instantiate the client. This ensures a much more secure but also reproducible code.

### Calling SAP Servers inside a VPC

By default, Lambda functions do not have access to private networks (a.k.a. VPCs). They can only access resources that have a public IP and are reachable from the internet (like my test server). There is a very easy fix: Lambda functions can be launched inside a VPC. They would then pull a free private IP address and use it for their networking operations. While this comes with some additional things to consider, it ensures that your data never traverses the public internet.