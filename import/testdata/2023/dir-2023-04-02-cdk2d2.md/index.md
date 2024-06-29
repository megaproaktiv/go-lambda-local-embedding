---
title: "A new simple approach to diagram as code on AWS with CDK and D2"
author: "Gernot Glawe"
date: 2023-04-02
draft: false
image: "img/2023/04/transformation-small.png"
thumbnail: "img/2023/04/transformation-small.png"
toc: true
keywords:
    - drawing
    - cdk
tags:
    - level-100
categories: [aws]

---

A diagram should convey a clear message about the intention of the architecture. For this message, you only need a few primary resources. Most generated diagrams are overloaded. This new app generates a diagram as code from your annotations in the CDK code.

<!--more-->


When I am doing AWS Training, participants often state the wish that diagrams are automatically drawn from AWS resources.
As this is possible with solutions like cfn-diag, the generated diagrams usually need more information to be helpful.
A diagram should convey a clear message about the architecture's intention. For this message, you only need a few primary resources. The other "glue" resources are necessary for the proper functions but not to show the logical view of a diagram.

The traditional approaches to this problem are:

1) Only use it if you have a small number of resources, like the CloudFormation drawing tool.

![Cloudformation](/img/2023/04/adot-starter-cfn.png)

2) Filter resources from the diagram afterwards, like `cfn-dia`

![cdf-diag](/img/2023/04/serverless-diag.png)

3) Draw manually

All solutions have a low degree of automation and do not synchronize with changes in the infrastructure.

A helpful solution should:

1) be fully automated, with no manual steps
2) Be 100% synched with deployed resources. Show *what is* and not what should be
3) Take meaningful names from the *Code* part of Infrastructure as Code. The often used CloudFormation logical ID is not helpful.

And my main requirement:

4) Show only those resources which I think are needed, to show the main purpose of the architecture

## A simple approach to DaC

1) Generate [D2](https://d2lang.com/tour/intro/) code from CDK code. D2 will take care of rendering.

2) Poll CloudFormation for the deployment status of the resource. Only resources which are beeing created or are created are shown.

3) Use the CDK Construct ID as title for a resources, not the logical ID or the physical ID.

4) Add very few information to the CDK code, to keep thinks simple:
  - `Show` - the Construct is shown on the diagram
  - `Connection` - a connection is drawn from source Construct to target Construct
  - `Contains` - Resources like a VPC, which contain other resources

## Example with a serverless architecture

I take [simple serverless architecture](https://github.com/megaproaktiv/adot-xraystarter) as a starting point.

These are the lines added for the diagramm information into the [TypeScript file](https://github.com/megaproaktiv/adot-xraystarter/blob/main/lib/xraystarter-stack.ts):

```ts
import * as d2 from "../lib/d2";
d2.Show(fnTS)
d2.Show(fnPy)
d2.Show(fnGO)
d2.Show(bucky)
d2.Show(topic)
d2.Connection(bucky, topic)
d2.Connection(topic, fnTS)
d2.Connection(topic, fnGO)
d2.Connection(topic, fnPy)
d2.Connection(fnTS, table)
d2.Connection(fnGO, table)
d2.Connection(fnPy, table)
```

We have a S3 Bucket named "bucky", some Lambda FNunctions, a SNS topic and a DynamoDB table.

And I copy  a small TypeScript [file](https://github.com/megaproaktiv/cdk2d2/blob/main/testdata/lib-ts/d2.ts) into the `lib` directory.

I deploy the infrastructure:

```bash
cdk deploy
```

Then I generate a d2 diagramm description with [cdk2d2](https://github.com/megaproaktiv/cdk2d2):

```bash
 cdk2d2 generate adotstarter-auto .
```

Where the first parameter is the *name* of the stack  and the second parameter is the *directory* of the CDK app. This creates a file `adotstarter-auto.d2`.

With d2 I generate a png of this file:

```bash
d2 adotstarter-auto.d2 adotstarter-auto.png
```

Now we got:

![cdn](/img/2023/04/adotstarter-auto.png)

instead of the quite confusing CloudFormation diagram:

![d2](/img/2023/04/adot-auto-cfn-highlight.png)

You see on first sight what the architecure is doing *and* the name of the Functions are not `adotstartergo2987B222`, which would be the CloudFormation *logical* ID, but the *Construct* ID from the code:

```ts
const fnGO = new aws_lambda.Function(this, 'adotstarter-go', {
```

## Inside cdk2d2

![dataflow](/img/2023/04/cdk2d2-dataflow.png)

With the information from the CDK manifest file from the [cloud-assembly](https://github.com/aws/aws-cdk/tree/v1-main/packages/%40aws-cdk/cloud-assembly-schema), q construct like the lambda function "fnGO" is generated as D2:


```d2
adotstartergo2987B222: adotstarter-go{
 icon: https://icons.terrastruct.com/aws%2FCompute%2FAWS-Lambda_Lambda-Function_light-bg.svg
 style.fill:"lightgreen"
}
```

Where the id of the artefact is the cfn logicalid and the title is the CDK construct ID.
The icon is set according to the CloudFormation Resource type. See [icon.go](https://github.com/megaproaktiv/cdk2d2/blob/main/monitor/icon.go) for mappings. The color `style.fill` is set according to the deployment state of the resource.

So when you change the name of the construct, the diagram is updated with the new name.

*You* define the resources which should be shown on the diagramm in your code, because only you know which resources are the main resources.
For example the helper lambda which CDK uses to set the retention time on the CloudWatch logs should usually now be shown.

## Now to the synchronisation part

I destroy the stack. Now we use three terminal windows:

1) cdk2d2 Script generator
2) d2 rendering
3) CDK deploy.

All running in the same CDK app base directory.

### Terminal 1

On terminal 1 we start:

```bash
cdk2d2 watch adotstarter-auto .
```

This will update the stackname.d2 each 4 seconds with then current deployment state of the construct resource.

### Terminal 2

On this terminal we start the rendering process with:

```bash
d2 --watch adotstarter-auto.d2
```

This will open a browser windows, where d2 renders the live state of the stack.de diagram.

### Terminal 3

Here we start the CDK deployment.

```bash
cdk deploy
```

### First picture

Will show nothing, because the stack does not exists.

### Second picture

Will show the stack in the upper left with the total number of resources and the deployment percentage.

![Stack only](/img/2023/04/cdk2d2-stack-only.png)

### Life state change

![Life 1](/img/2023/04/life-1.png)

Resources which are beeing created are colored yellow.

![Life 2](/img/2023/04/life-2.png)
Resources which are created are colored green.

### Fully deployed state
![Life 3](/img/2023/04/life-3.png)

100% resources created.



For each resource you will see the deployment status as color.
This way you can watch CloudFormation deploying your resources. The order of the deployment will surprise you sometimes...

## Dependencies

The metainformation are stored as metadata:

### Show

`resource.node.addMetadata("Show", "true")`

### Connection

`resource.node.addMetadata("Connection", otherResource.node.id)`

### Container

`resource.node.addMetadata("Container", otherResource.node.id)`


## Container example

Of course we are talking about a graphical container!

![VPC](/img/2023/04/application-load-balancer.png)

For a VPC you add a "Container" relationship. Only one level is supported, so vpc-subnet will not work.


## Now try it out for yourself

Please send suggestions, errors, images for the gallery etc. Add prs for new features to [cdk2d2](https://github.com/megaproaktiv/cdk2d2) issues. There are also some more examples. Releases can be found [here](https://github.com/megaproaktiv/cdk2d2/releases).



If you like it, please star the repo!

Happy building!


If you need consulting for your serverless project, don't hesitate to get in touch with the sponsor of this blog, [tecRacer](https://www.tecracer.com/kontakt/).

For more AWS development stuff, follow me on dev https://dev.to/megaproaktiv.
Want to learn GO on AWS? [GO here](https://www.go-on-aws.com/)

## See also

- [cdk2d2 Source code ](https://github.com/megaproaktiv/cdk2d2)
- [d2](https://d2lang.com/tour/intro/)
- [AWS CDK](https://aws.amazon.com/cdk/)
