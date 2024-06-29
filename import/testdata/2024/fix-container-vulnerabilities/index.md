---
title: "From fragile to formidable: How to detect, fix and prevent container vulnerabilities with Inspector and Docker Scout"
author: "Gernot Glawe"
date: 2024-02-22
draft: false
image: "img/2024/02/fcv/camilo-jimenez-vGu08RYjO-s-unsplash.jpg"
thumbnail:  "img/2024/02/fcv/camilo-jimenez-vGu08RYjO-s-unsplash.jpg"
toc: true
keywords:
    - inspector
    - docker
    - go
    - nginx
    - python
tags:
    - level-300
    - security

categories: [aws]
---

A webserver running on a container. Sound simple. Let`s dive deeper into how your architecture choices affect application security. I use docker scout for the container and show how Amazon Inspector can serve as a general-purpose security tool.

<!--more-->

The first step is to reduce the attack surface. This paradigm works contrary to the DRY - do not repeat yourself - principle from current scripting languages.If you use a powerful framework for you application, developing is faster and easier. But the framework is a huge attack surface. If you use a small library, you have to write more code, but the attack surface is smaller.

There are tradeoffs you should be aware of. In this post, I show you how to reduce attack surface and go from a fragile architecture to a formidable one.

## Architecture overview

![](/img/2024/02/fcv/architecture.png)

### Level 0 -  Host OS

If you have the possibility to use fargate, do it! You gain a lot of security. The tradeoff is that you gain more control, if you use EC2 instances. E.g. just ssh into the instance and look at the logs. Amazon Inspector can be used on EC2 instances. But thats a story for another post.

### Level 1 - Docker

Inspector inspects container. It this test I found additional vulnerabilities, which were not detected with inspector, with docker scout.

### Level 2 - Interpreter

The interpreter itself can have vulnerabilities. The border between the interpreter and the base libraries is fluent.

On AWS you can automate security patching, if you replace container with Lambda functions. The intereter runtimes are patched by AWS on Lambda.


### Level 3 - Dependencies

With each library you use, the **attack surface** grows. The more libraries you use, the more likely it is that one of them has a vulnerability.
Scanning before deploying is not enough. Some vulnerabilities are only detected after the code is running. Amazon Inspector does continuous scanning.

### Level 4 - Application

You can code in an unsecure way. The code vulnerabilitie problems will grow in the future with AI generated code. That discussion is not part of this post.

## From fragile to formidable

The wording comes from an llm, but I find it very fitting for the topic.

## Easy to implement, but fragile: Streamlit Python app

As described in [this blogpost from Dec 23](/2023/12/go-ing-to-production-with-bedrock-rag-part-1.html) a easy to implement POC architecture is implemented with [streamlit](https://streamlit.io/).

This is the first attempt for the architecture:

![](/img/2024/02/fcv/architecture-fragile.png)

The Dockerfile is simple:

```Dockerfile
# app/Dockerfile

FROM --platform=linux/amd64 python:latest

WORKDIR /app

RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt work/

WORKDIR /app/work
RUN pip install --no-cache-dir -r requirements.txt

WORKDIR /app
RUN mkdir work/images
COPY ./*.py work/
COPY ./*.yaml work/

WORKDIR /app/work

EXPOSE 8501

HEALTHCHECK CMD curl --fail http://localhost:8501/_stcore/health

ENTRYPOINT ["streamlit", "run", "app.py", "--server.port=8501", "--server.address=0.0.0.0"]
```

To run a simple streamlit app, you need a lot of dependencies.

Example `requirements.txt`:

```txt
altair==5.2.0
attrs==23.1.0
bcrypt==4.1.1
blinker==1.7.0
boto3==1.34.29
botocore==1.34.29
cachetools==5.3.2
certifi==2023.11.17
charset-normalizer==3.3.2
click==8.1.7
extra-streamlit-components==0.1.60
gitdb==4.0.11
GitPython==3.1.40
idna==3.6
importlib-metadata==6.11.0
Jinja2==3.1.2
jmespath==1.0.1
jsonschema==4.20.0
jsonschema-specifications==2023.11.2
markdown-it-py==3.0.0
MarkupSafe==2.1.3
mdurl==0.1.2
numpy==1.26.2
packaging==23.2
pandas==2.1.3
Pillow==10.1.0
protobuf==4.25.1
pyarrow==14.0.1
pydeck==0.8.1b0
Pygments==2.17.2
PyJWT==2.8.0
python-dateutil==2.8.2
pytz==2023.3.post1
PyYAML==6.0.1
referencing==0.31.1
requests==2.31.0
rich==13.7.0
rpds-py==0.13.2
s3transfer==0.10.0
six==1.16.0
smmap==5.0.1
streamlit==1.29.0
streamlit-authenticator==0.2.3
tenacity==8.2.3
toml==0.10.2
toolz==0.12.0
tornado==6.4
typing_extensions==4.8.0
tzdata==2023.3
tzlocal==5.2
urllib3==2.0.7
validators==0.22.0
zipp==3.17.0
```

To deploy a container in ECR (Elastic Container Registry), I use this Taskfile:

```yaml
# https://taskfile.dev

version: "3"

vars:
  REGION: "eu-central-1"
  NAME: "fragile"
  VERSION: "v0.1.1"
  ACCOUNT:
    sh: aws sts get-caller-identity --query Account --output text

tasks:
  ecr-push:
    desc: Build local image and Push to ECR, docker or colima must run before
    cmds:
      - task: docker-build
      - task: docker-tag
      - task: docker-login
      - task: docker-push

  docker-build:
    desc: Docker build (1)
    cmds:
      - docker build -t {{.NAME}} .

  docker-tag:
    desc: Docker tag (2)

    cmds:
      - docker tag {{.NAME}}:latest {{.ACCOUNT}}.dkr.ecr.{{.REGION}}.amazonaws.com/{{.NAME}}:{{.VERSION}}

  docker-login:
    desc: Docker tag (3)

    cmds:
      - aws ecr get-login-password --region {{.REGION}} | docker login --username AWS --password-stdin {{.ACCOUNT}}.dkr.ecr.{{.REGION}}.amazonaws.com

  docker-push:
    desc: Docker tag (4)
    cmds:
      - docker push {{.ACCOUNT}}.dkr.ecr.{{.REGION}}.amazonaws.com/{{.NAME}}:{{.VERSION}}

  create-ecr:
    desc: Create ECR
    cmds:
      - aws ecr create-repository --repository-name {{.NAME}} --region {{.REGION}} --image-scanning-configuration scanOnPush=true
```

For the other architectures, you replace the `NAME` variable. Fun fact: It speaks for AWS stability, that I used these steps [6 years ago](https://github.com/tecracer/demo-fargate) and they still work. I only traded `Makefile` for `taskfile.dev`.


**Steps to deploy:**

1) `task create-ecr`
2) `task ecr-push`

**Scanning Prequesites:**

1) Activate Inspector on the AWS account

### Fragile Vulnerabilities



The inspector is activated and the image is scanned:


![](/img/2024/02/fcv/fragile-ecr.png)

OOPS, 2 critical vulnerabilities and 314 other.

Let's analyse:

#### Level 1 - Base Image

With docker scout you get some insights about the base image:

```bash
docker scout quickview local://fragile
    ✓ SBOM of image already cached, 629 packages indexed

  Target             │  local://fragile:latest  │    0C     4H     8M   100L     1?
    digest           │  a4a879c6f9a8            │
  Base image         │  python:3                │    0C     2H     7M    99L
  Updated base image │  python:3-slim           │    0C     1H     1M    21L
                     │                          │           -1     -6    -78
```

So a first step towards more security would be to use a slim version of the base image.

#### Level 2 - Python

![](/img/2024/02/fcv/fragile-ecr-filter-1.png)

Lets look at the details of the first vulnerability:


Name | Package | Severity |Description
:----|:--------|:---------|:-----------
CVE-2023-24329	|python3.11:3.11.2-6 |	HIGH	|An issue in the urllib.parse component of Python before 3.11.4 allows attackers to bypass blocklisting methods by supplying a URL that starts with blank characters.

In the text you see, in which version the bug is fixed. What you also have to check is whether you really _use_ the package in your code. If not, you can delay fix of the vulnerability.

To further analyse the vulnerabilities, you can click the [Name: CVE-2023-24329](https://security-tracker.debian.org/tracker/CVE-2023-24329) i the AWS console.

You see many things. Notworthy is, that using a new version of the interpreter is not always a way to fix the vulnerability:

![](/img/2024/02/fcv/cve-1.png)

So use a newer version, but keep scanning :) .

## Sturdy - A more secure version

This example architecture splits the monolithic fragile architecture  into a frontend and a backend. The frontend is a static website, serverd by nginx. For simplicity we assume the backend is written as Lambda. So its a bit unfair to compare the monolithic archicture, which contains front- and backend. Its only done to show the difference in the vulnerabilities and to have a good story ;) .

![](/img/2024/02/fcv/architecture-sturdy.png)

As an example, the Dockerfile just has nginx as base image:

```Dockerfile
FROM nginx
```

The hope is to get less vulnerabilities and the latest nginx version.

### Sturdy Vulnerabilities

```bash
docker scout quickview local://sturdy
    ✓ Image stored for indexing
    ✓ Indexed 232 packages

  Target             │  local://sturdy:latest  │    0C     1H     2M    37L
    digest           │  760b7cbba31e           │
  Base image         │  debian:12-slim         │    0C     0H     0M    18L
  Updated base image │  debian:stable-slim     │    0C     0H     0M    18L
```

Inspector also has some things to say:

![](/img/2024/02/fcv/sturdy-ecr.png)

As Dante says: "Lasciate ogne speranza, voi ch'intrate" - "Abandon all hope, ye who enter here".

So using a standard webserver without optimization does not help to reduce vulnerabilities.

## Potent - build your own server

![](/img/2024/02/fcv/architecture-potent.png)


One road to zero vulnerabilities is to build your own server.
The paradign shifts from "use frameworks and libraries" to "use as few libraries as possible".

The Dockerfile is simple. The GO build itself is done independent of the base image dockerfile.

```Dockerfile
FROM alpine:latest
WORKDIR /app
COPY dist/server_linux server

EXPOSE 80
ENTRYPOINT ["./server"]
```

The base GO code uses a framework ([Gin](https://gin-gonic.com/)), you could also use the standard library. But as the scan shows, that is not needed as we are already have less vulnerabilities.

The main code is simple:

```go
router := gin.Default()

// Middleware to protect sensitive routes

// Serve static files
authorized := router.Group("/", basicAuth)

authorized.StaticFile("/", "./public/index.html")
authorized.StaticFile("/asset-manifest.json", "./public/asset-manifest.json")
authorized.StaticFile("/manifest.json", "./public/manifest.json")
authorized.StaticFile("/favicon.ico", "./public/favicon.ico")
authorized.StaticFile("/index.html", "./public/index.html")
```

Here the app *only* serves single files, not all files. So the attack surface is reduced.

### Potent Vulnerabilities

![](/img/2024/02/fcv/potent-ecr.png)

Because we used the minimal linux "alpine" the Level 1 vulnerabilities are (almost) gone.

With a static compiled binary, we do not need so much operating system libraries. So the **attack surface** is *reduced*.

Amazon Inspector directly shows the problem _and_ the solution:

![](/img/2024/02/fcv/potent-ecr-filter-1.png)

So update the library and we have *zero* vulnerabilities:

![](/img/2024/02/fcv/potent-zero.png)

Time to party? Not yet. Scout shows vulnerabilities in the base image `alpine:3`.



![](/img/2024/02/fcv/potent-not-zero.png)

Time for the last mile.

## Formidable - Zero vulnerabilities

![](/img/2024/02/fcv/architecture-formidable.png)

```dockerfile
FROM alpine:3.19.1 as builder
WORKDIR /app
COPY dist/server_linux server


EXPOSE 80
ENTRYPOINT ["./server"]
```

## ECR says zero vulnerabilities

![](/img/2024/02/fcv/formidable-ecr.png)

Scout is also happy:

![](/img/2024/02/fcv/formidable-scout.png)

*NOW* it is time to party!

## Summary

We learn two things:

Amazon Inspector is a great tool to find vulnerabilities not only in your Docker images. But for *zero vulnerabilities* you have to go the extra mile and use additional tools like [Docker Scout](https://dockerscout.com/).

The second thing is, that powerfull frameworks trade security for features. The **attack surface** is larger, so is the probability of vulnerabilities.

## What's next?

I do not talk about application induced vulnerabilities. Look for scanners in you development language. For example, in Go you can use [govulncheck](https://go.dev/blog/vuln).


If you need AWS training, developers and consulting for your next cloud project, don't hesitate to contact us, [tecRacer](https://www.tecracer.com/kontakt/).

Want to learn GO on AWS? [GO here](https://www.udemy.com/course/go-on-aws-coding-serverless-and-iac/learn/lecture/38434460?referralCode=954E43527F32E22BB1C7#overview)

Enjoy building!

## Thanks to
Photo by <a href="https://unsplash.com/@camstejim?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">camilo jimenez</a> on <a href="https://unsplash.com/photos/red-vehicle-in-timelapse-photography-vGu08RYjO-s?utm_content=creditCopyText&utm_medium=referral&utm_source=unsplash">Unsplash</a>

- icon by Pham Duy Phuong Hung
