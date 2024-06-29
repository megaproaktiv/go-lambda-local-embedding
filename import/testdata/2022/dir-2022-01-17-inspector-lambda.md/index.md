---
title: "Enhance Lambda Security with new Amazon Inspector Vulnerability Management and prevent 'log4jgate'"
author: "Gernot Glawe"
date: 2022-01-14
draft: false
image: "img/2022/01/IMG_5204-lambda2.png"
thumbnail: "img/2022/01/IMG_5204-lambda2.png"
toc: true
keywords:
    - security
    - inspector
    - vulnerability
    - level-300
    - cdk
tags:
    - level-300
    - lambda
    - cdk
categories: [aws]

---

Detect the crack in the window (or the lambda library) before it breaks:

As we have seen during the last month, also well known libraries like [log4j](https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=log4j) can have previously unknown vulnerabilities.

Therefore scanning your Lambda application before deploying is not enough! What happens if a new cybersecurity vulnerability occurs while your functions are running? Solution: Amazon Inspector. Only problem: it`s not available for Lambda yet. Fortunately, you can deploy Lambda with container images and Inspector will continuously scan these images for you. 

 Want to know how set up Lambda & Inspector and see how evil Node vulnerabilities are detected? Read on!


<!--more-->

## The reason for continuous vulnerability scanning

Lambda is often used in web applications. To handle threads you model the threads. That means you think about all directions attacker may come from.

A thread model for a simple Lambda application may look like this:

![Demo Model](/img/2022/01/2022-01-14_15-42-42.png)

A vulnerability is a weakness in the application, here the used libraries could be exploited by threads.

So a vulnerability is like a window with a crack. When nobody punches, everything is ok. But if an attack occurs, the window will break.

The problem is that these weaknesses in the library can be detected **after** you deploy your application. If you just scan with the deployment once, these vulnerabities will remain undetected.

Nobody knows beforehand, which library includes a weakness. But as people use libraries and analyze code, the weaknesses are detected and are stored in databases. An example is the `mattermost` library in the following examples, shown in the snyk database: [Snyk](https://security.snyk.io/vuln/SNYK-JS-MATTERMOSTMOBILEE2E-2323399).


## New Amazon Inspector

Amazon Inspector is a vulnerability management service that **continuously** scans your AWS workloads for vulnerabilities. Amazon Inspector automatically discovers and scans Amazon EC2 instances and container images residing in Amazon Elastic Container Registry (Amazon ECR) for software vulnerabilities and unintended network exposure. Inspector uses several vulnerability databases.

The new service is discussed in detail in the great AWS podcast : [#491: INTRODUCING The New Amazon Inspector](https://aws.amazon.com/podcasts/491-introducing-the-new-amazon-inspector/)

Another feature is the contextualized risk score. That means, you not only get a "HIGH" risk score for the findings, but inspector takes the AWS context in the risk score calculation

See the [AWS news blog](https://aws.amazon.com/blogs/aws/improved-automated-vulnerability-management-for-cloud-workloads-with-a-new-amazon-inspector/?trkCampaign=AWS_Podcast_491_Podcast&sc_channel=el&sc_campaign=SM_2021_AWSWNL_491&sc_outcome=Launch_Marketing)

The feature I will use here is: **support for container-based workloads**

I wanted to answer the question: Can inspector be used for scanning lambda applications? The answer is: "Yes, indeed"

## Lambda Container

## From Lambda Deployment to container deployment

Usually, the lambda packages are imported as files or packages directly to lambda. With the [container image](https://docs.aws.amazon.com/lambda/latest/dg/lambda-images.html) feature, you create an image, upload it to the container registry ECR and deploy it to lambda.

### Pros:

-  Local Lambda execution right out of the box!

    See [Testing Lambda container images locally](https://docs.aws.amazon.com/lambda/latest/dg/images-test.html) for further detail. You do not have to use additional tooling like AWS SAM or lambci.


- Larger image size possible

    The size for a lambda deployment package is 250 Megabyte unzipped. See your AWS quotas page for details. With container images, this increases to 10 GB!

    See [Creating Lambda container images](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html).


### Cons

- Slower Deployment
- You have to take care of the base image yourself. 
- Slightly Slower cold start


Now let us build some lambda container images:

## Dockerfile for Node/GO/Python/Java Lambda images


### Node

```dockerfile
FROM public.ecr.aws/lambda/nodejs:14

COPY index.js package.json  ${LAMBDA_TASK_ROOT}

RUN npm install

CMD [ "index.handler" ]  
```

This is a simple Dockerfile, which only copy *one* file, the `index.js`. All files have to be copied into the LAMBDA_TASK_ROOT. which is set by the CDK.
You could tweak the `npm install` to exclude development dependencies.

### Python

```dockerfile
FROM public.ecr.aws/lambda/python:3.8

COPY app.py ${LAMBDA_TASK_ROOT}

COPY requirements.txt  .
RUN  pip3 install -r requirements.txt --target "${LAMBDA_TASK_ROOT}"

CMD [ "app.handler" ] 
```

The basic steps are:

- Copy the application itself
- Install dependencies
- Run handler

The advantage of using this approach for Python is, that you could compile operation system dependant libraries directly with Linux. Because lambda also runs on Linux.

### GO

```dockerfile
FROM public.ecr.aws/lambda/provided:al2 AS build
ENV CGO_ENABLED=0
RUN mkdir -p /opt/extensions
RUN yum -y install go 
RUN go env -w GOPROXY=direct
ADD go.mod go.sum ./
RUN go mod download
COPY . ${LAMBDA_TASK_ROOT}
RUN env GOOS=linux GOARCH=amd64 go build -o=/main 
# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /main /main
ENTRYPOINT [ "/main" ]
```

Here the steps are different, because GO builds a static binary:

- Install dependencies: `go mod download`
- Build the app: `go build`
- Create a clean image only with the application binary

### Java

```dockerfile
FROM public.ecr.aws/lambda/java:11

COPY target/classes ${LAMBDA_TASK_ROOT}
COPY target/dependency/* ${LAMBDA_TASK_ROOT}/lib/

CMD [ "demo.lambda::handleRequest" ]
```

Here the java application was compiled with maven:

```bash
mvn compile dependency:copy-dependencies -DincludeScope=runtime
```

This lets all classes be compiled to `target/classes` and all libraries to `target/dependency/`
The steps are defined in the maven `pom.xml`, see the source file for that.

Here the java application was build outside the Dockerfile, so that the precompiled code is stored before the image is build.

## Deploy with CDK

The container deployment is simplified with the CDK. We use the Lambda Construct "Docker Image Function".

### CDK Typescript Lambda Container

In TypeScript CDK code you just need these lines and a running Docker on your laptop to deploy Lambda container images:

```typescript
  var dockerfile = path.join(__dirname, '../../app/node')
    const dockerlambda = new aws_lambda.DockerImageFunction(this, "lambdainspector-node",
    {
      architecture: aws_lambda.Architecture.X86_64,
      functionName: "lambdainspector-node",
      timeout: Duration.seconds(1024),
      code: aws_lambda.DockerImageCode.fromImageAsset(dockerfile)
    })
```

### GO Lambda Container

The same CDK code in GO:

```go
lambdaArchitecture := awslambda.Architecture_X86_64()

dockerfile = filepath.Join(path, "../app/node")
awslambda.NewDockerImageFunction(stack,
    aws.String("lambdainspector-node"),
    &awslambda.DockerImageFunctionProps{
        Architecture:                lambdaArchitecture,
        FunctionName:                 aws.String("lambdainspector-node"),
        MemorySize:                   aws.Float64(1024),
        Timeout:                      awscdk.Duration_Seconds(aws.Float64(300)),
        Code:                         awslambda.DockerImageCode_FromImageAsset(&dockerfile, 
        &awslambda.AssetImageCodeProps{}),
    })

```        

For the other Lambda languages of the application (not the Lambda Resource, which is the infrastructure), you just point the `dockerfile` path to another location.

```go
	dockerfile = filepath.Join(path, "../app/python")
	awslambda.NewDockerImageFunction(stack,
		aws.String("lambdainspector-py"),
		&awslambda.DockerImageFunctionProps{
			Architecture:                 lambdaArchitecture,
			FunctionName:                 aws.String("lambdainspector-py"),
			MemorySize:                   aws.Float64(1024),
			Timeout:                      awscdk.Duration_Seconds(aws.Float64(300)),
			Code:                         awslambda.DockerImageCode_FromImageAsset(&dockerfile, &awslambda.AssetImageCodeProps{}),
		})
```

See [go-on-aws Lambda Container Deployment](https://www.go-on-aws.com/lambda-go/deploy/deploy_lambda_container_amd64/) for details.

## Deploy images

You may activate inspector before you deploy images or afterwards. The continuous scanning will trigger the scanning process either way.

When you clone the repository, you change to the `lambda-inspector` base directory.

The different lambda are stored in `app`, the CDK application is in the directory `infra`.

So deploy the images with:

```bash
cd lambda-inspector/infra
npx cdk@2.7.0 deploy
```

For the node app and the java app I have included vulnerabilities:

## Call the evil Ghost

See the evil [package.json](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-inspector/app/node)

```json
{"dependencies": {
    "aws-sdk": "^2.1055.0",
       "@fabiocaccamo/utils.js": "^0.17.1",
    "cordova-plugin-fingerprint-aio": "^5.0.0",
    "discordjs-lofy": "^0.0.1-security",
    "istanbul-reports": "^3.1.2",
    "jquery.terminal": "^1.20.4",
    "js-data": "^3.0.10",
    "jsuites": "^4.9.10",
    "jsx-slack": "^4.5.0",
    "mattermost-mobile": "^0.0.1-security",
    "mattermost-mobile-e2e": "^0.0.1-security",
    "mermaid": "^0.2.11",
    "momnet": "^1.12.1",
    "parse-link-header": "1.0.1"
    }
}
```

All these packages have known weaknesses. DO NOT USE THEM, this is just to test the Amazon Inspector!

The code in the Dockerfile will install the packages:


```dockerfile
# Install NPM dependencies for function
RUN npm install
```

Lets see if inspector will find anything.

## Who you gonna call?

Inspector... 

### Before evil

With a clean `package.json` all Lambdas are safe, no findings:

![](/img/2022/01/2022-01-12_08-04-54.png)

### The CDK Repo

CDK will create an own ECR repository for all images:

![](/img/2022/01/2022-01-12_08-06-46.png)

### Scanning the evil package.json

After deploying the affected containers, these are the findings for the Node app:

![](/img/2022/01/2022-01-12_08-42-39.png)

![](/img/2022/01/2022-01-12_08-42-19.png)



## Going nuclear - log4j

![](/img/2022/01/2022-01-14_14-15-45.png)

To test the vulnerability hype these days, I add a Lambda with java and Log4.

The result: immediately after pushing the image, Inspector detected the vulnerabilities!

## Ghostbusting

To find the lambda code where the bad libs are used, we have to do the following steps:

![](/img/2022/01/2022-01-14_14-18-11.png)

1) In the inspector console, find the CDK repository
2) Get the image tag, here for the first image a6cf...
3) Find the tag in the generated CloudFormation template, e.g. with `grep`

```bash
grep a6cf2f65b84e9ea851f8674454a6c9388bcc928909d380015e56d76fd3ed36d5      cdk.out/LambdaInspectorStack.template.json -B 5
    "lambdainspectornode000CEAD7": {
      "Type": "AWS::Lambda::Function",
      "Properties": {
        "Code": {
          "ImageUri": {
            "Fn::Sub": "${AWS::AccountId}.dkr.ecr.${AWS::Region}.${AWS::URLSuffix}/cdk-hnb659fds-container-assets-${AWS::AccountId}-${AWS::Region}:a6cf2f65b84e9ea851f8674454a6c9388bcc928909d380015e56d76fd3ed36d5"
--
      "DependsOn": [
        "lambdainspectornodeServiceRoleD8267A01"
      ],
      "Metadata": {
        "aws:cdk:path": "LambdaInspectorStack/lambdainspector-node/Resource",
        "aws:asset:path": "asset.a6cf2f65b84e9ea851f8674454a6c9388bcc928909d380015e56d76fd3ed36d5",
```

As CDK stores all CloudFormation templates in `cdk.out`. The name of the function is shown above the `"Type"` line:

```json
{
"lambdainspectornode000CEAD7": {
      "Type": "AWS::Lambda::Function",
}
```

Now we can find the function in the CDK app:

```go
	dockerfile = filepath.Join(path, "../app/node")
	awslambda.NewDockerImageFunction(stack,
		aws.String("lambdainspector-node"),
        FunctionName:                 aws.String("lambdainspector-node"),
        ---
		})

```        

And update the function libraries!


## Summary

1) Inspector does continuous scanning well and in time, which provides much better security
2) Lambda native deployment is not supported yet
3) You can use Lambda Images to get continuous scanning

No animals were harmed during the blog post, the fly was dead before the photo. 😉

## See the full source on github.

- [Sources](https://github.com/megaproaktiv/aws-community-projects/tree/main/lambda-inspector)


## Feedback & discussion

For discussion please contact me on twitter @megaproaktiv

## Learn more GO

Want to know more about using GOLANG on AWS? - Learn GO on AWS: [here](https://www.go-on-aws.com/)

