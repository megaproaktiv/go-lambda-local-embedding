---
author: "Thomas Heinen"
title: "Docker Architecture - Intel? ARM? both?"
date: 2023-03-16
image: "img/2023/03/gpt4-midjourney5_b6e417df-a5cc-4c25-ada4-2dd6e237d15e_1280.png"
thumbnail: "img/2023/03/gpt4-midjourney5_b6e417df-a5cc-4c25-ada4-2dd6e237d15e_1280.png"
toc: false
draft: false
tags: ["docker", "vscode", "level-200", "arm", "graviton"]
---
Up to a few years back, writing Dockerfiles was easy. In many cases, it still is - unless you are working with a mixed fleet of Intel and ARM-based processors. Are you familiar with this situation and you do not want to maintain two almost identical Dockerfiles? There is a solution...

<!--more-->

## History and Motivation

CPU Architectures were mostly homogenous for a while. After Apple [switched from its previous PowerPC platform to Intel-based Macs in 2006](https://en.wikipedia.org/wiki/Mac_transition_to_Intel_processors), the majority of the desktop and servers were running on the familiar "x86" architecture.

Recently, this changed - and it keeps changing. We now have the ARM-based M1/M2 processors in modern Apple computers, embedded devices have been running ARM for years and years, and AWS iterates on their ARM processor lineup [Graviton](https://aws.amazon.com/ec2/graviton/), after [acquiring Annapurna Labs in 2015](https://www.crunchbase.com/acquisition/amazon-acquires-annapurna-labs--34f1f987). Currently at "Graviton 3", these CPUs offer a lot of computing power for little money. And they are much more energy-efficient, making up a good part of the [Well Architected Framework](https://docs.aws.amazon.com/wellarchitected/latest/framework/welcome.html)'s Sustainability pillar.

There is even another architecture on the horizon, namely the [RISC-V architecture](https://en.wikipedia.org/wiki/RISC-V). While also being in the same RISC architecture family as ARM, it does not have licensing issues involved. This means, that any company can use the design and create chipsets to distribute freely.

This revival of a multi-architecture landscape also creates problems - like the aforementioned platform conflicts on servers or even desktops. With the prevalence of Docker, you now need to cover two (or even more) platforms at the same time. While most Linux distributions will be available for all of them and have identical commands and packages, you will have problems if you rely on standalone programs, especially on GitHub.

## Theory

Theoretically, the distinction between these two architectures should be simple. You would expect a command to output the architecture, in a distribution-agnostic way. And then, downloads on GitHub should just have that in their package name for easy access.

Of course, that's not the case.

In essence, there are at least two schools of thought
- Intel can be referred to as `i386` (outdated 32 bit), `i686`, `x86_64` or `amd64`
- Graviton/Apple Silicon can be referred to as `aarch64`, `arm`, or `arm64`

And of course, there are all those other architectures from embedded systems (`armhf`, `armel`), mainframes (`s390), and exotic architectures (`powerpc`, `sparc`).

Luckily, popular packages seem to fall into two categories:
- they offer downloads with `x86_64` and `aarch64` in the package names
- or they offer them as `arm64` and `amd64`

A few solutions are mixing both styles. Therefore, we will need to address this (later).

## Alternative 1: Build Arguments

Docker offers the possibility to pass in arguments at build time. So you can just assign variables for your architectures and use those in the Dockerfile:

```Dockerfile
FROM ubuntu:22.04

ARG ARCH
ARG ARCH_ALT

# ...

RUN curl https://packages.chef.io/repos/apt/stable/ubuntu/20.04/inspec_5.21.29-1_$ARCH.deb --output /tmp/inspec.deb --silent \
    && dpkg --install /tmp/inspec.deb \
    && rm /tmp/inspec.deb
```

When you now build your Docker image, you will do this via

```shell
docker build --build-arg ARCH="amd64" .
docker build --build-arg ARCH="arm64" .
```

This makes the whole process straightforward in your pipelines and keeps the `Dockerfile` clean. On the flip side, it will not work in situations where the image is built on-the-fly.

## Alternative 2: Embedding Commands

There are some commands which are commonly used to determine the system architecture:

- `dpkg --print-architecture` will result in the `amd64`/`arm64` pair, but is only available on Debian/Ubuntu
- `uname -m` or `arch` will result in the `x86_64`/`aarch64` combination and is distribution-agnostic

You can either use those in your `docker build` command like before or embed it in the `Dockerfile` itself:

```Dockerfile
FROM ubuntu:22.04

# ...

RUN curl https://packages.chef.io/repos/apt/stable/ubuntu/20.04/inspec_5.21.29-1_$(dpkg --print-architecture).deb --output /tmp/inspec.deb --silent \
    && dpkg --install /tmp/inspec.deb \
    && rm /tmp/inspec.deb
```

Unfortunately, this results in higher maintenance effort for your code. Until the time when Docker allows dynamic setting of environment variables, this is probably the best way to create Docker images if your influence on the build process is limited.

## Exceptions from the Rule

Some software packages cross their streams, as previously mentioned. In these cases, you will probably need some `if` clause in that particular `RUN` statement. Luckily, I have only encountered a handful of these.

```Dockerfile
FROM ubuntu:22.04

# ...

RUN ARCH=$(uname -m) && if [ "$ARCH" == "aarch64" ]; then ARCH="arm64"; fi \
    && curl --location-trusted https://github.com/charmbracelet/glow/releases/download/v1.5.0/glow_1.5.0_linux_$ARCH.tar.gz --output /tmp/glow.tgz --silent \
    && tar zxf /tmp/glow.tgz \
    && mv glow /usr/local/bin
```

## VSCode DevContainers

My main use case for this are [VSCode DevContainers](https://code.visualstudio.com/docs/devcontainers/containers). These allow users of VSCode or [GitHub Codespaces](https://github.com/features/codespaces) to package their development environment along with the code in their repositories - resulting in predictable environments for all participants in a project. They are a very effective way of avoiding the "Works on my machine" syndrome - and are based on Docker.

As you open a project in its DevContainer for the first time, VSCode will kick off an initial Docker image build process. Currently, that does not involve any dynamic elements or [pre-defined environment variables](https://containers.dev/implementors/json_reference/#variables-in-devcontainerjson) to use as architecture switches. As a result, we seem to be stuck with the second alternative: embedding commands.

Despite the additional maintenance issues, this has proven to work very well. Both my Intel-based colleagues and the Apple aficionados can now work inside identical development configurations.

For example, when writing new blog posts inside of their VSCode environment like I did with this post.
