---
author: "Thomas Heinen"
title: "Test-Kitchen on AWS (2022 edition)"
date: 2022-11-02
image: "img/2022/11/sd15-563811029.png"
thumbnail: "img/2022/11/sd15-563811029.png"
toc: false
draft: false
tags:
  - devops
categories:
  - AWS
  - Chef
  - Test Kitchen
---
Test-Kitchen is a tool to manage your test machine lifecycle, similar to HashiCorp Vagrant. While it has been developed with Chef in mind, it can be used with any development tool to test on new machines every time you change your code.

As this tool continues to evolve and many examples are outdated, today I will give you some small snippets to reuse and get going quickly.

<!--more-->

## Installation

There are two ways to install Test-Kitchen:

1. If you already have Ruby installed: `gem install test-kitchen kitchen-ec2`
2. Use the bundled installer with [Chef Workstation](https://www.chef.io/downloads/tools/workstation)

You can find more details in the [official documentation](https://kitchen.ci/docs/getting-started/introduction/). This article will focus on how to use it with AWS.

## Minimal Config

You will find much-outdated information if you search the Internet about using Test-Kitchen with AWS. For example, you are often asked to create IAM Profiles, SSH Keys, or Security Groups.

- IAM Profiles are not necessary unless you specifically want to use private S3 buckets etc
- SSH Keys will be automatically created and removed (your AWS user will, of course, need the corresponding privileges)
- Security Groups will also be automatically managed
- You do not need to find the `subnet-xxxxxxxx` ID to deploy your instance anymore, but you can use tagging instead

An essential addition is to switch to Instance Metadata Services v2 (IMDSv2). IMDS is responsible for providing access to the meta information of your EC2 instance, which could leak sensitive data with the older version 1. With the new IMDSv2, it is much harder to exploit it.

The following `kitchen.yml` file uses all these features:

```YAML
---
driver:
  name: ec2
  subnet_filter:
    tag: 'Name'
    value: '*public*'
  metadata_options:
    http_tokens: required
    http_put_response_hop_limit: 1
    instance_metadata_tags: enabled
  instance_type: t3a.medium
  associate_public_ip: true
  interface: public
  skip_cost_warning: true
  tags:
    CreatedBy: test-kitchen

platforms:
 - name: amazon2
 - name: amazon2-arm64
   driver:
     instance_type: t4g.nano
```

The `subnet_filter` assumes that you have Subnets with the `Name` tag, including the term `public`. Careful: it is not possible to specify a VPC so take care that you pick the right one if your AWS account includes multiple VPCs. Otherwise, TK will pick the first one returned - which might cause some problems.

With our `metadata_options`, we configure IMDS v2. The hop count of `1` should work for your applications; it usually only needs changing if you run containers. With the addition of `instance_metadata_tags` we can also use EC2 tags as metadata. If you use Chef, these will be available under `node['ec2']['tags_instance_MyTagName']`.

Notice under `platforms` that you can add a specific architecture like ARM64 for Graviton, which is possible since [version 2.3.3](https://github.com/test-kitchen/kitchen-ec2/commit/dba6a96596632afa44b88f7fc9615aba7468c7f6)

This architecture will then create EC2 instances in your public subnet, assign a temporary Security Group and let you SSH/WinRM into them for development.

<!-- Fun Blog: mac1.metal/mac2.metal with test-kitchen -->

## Avoiding SSH/WinRM

Newer AWS architectures recommend avoiding public instances and SSH, instead relying on AWS SystemManager (AWS SSM). This service works with a pre-installed agent on the most common AMIs and offers a lot of functionality, such as command execution, inventory, patching, and much more.

You must explicitly attach the corresponding permissions to your instance to enable it. If you want to do this via code, the following Terraform snippet creates a role and an EC2 profile to use:

```hcl
resource "aws_iam_role" "testkitchen" {
  name                = "test-kitchen"
  assume_role_policy  = data.aws_iam_policy_document.instance_assume_role_policy.json
  managed_policy_arns = ["arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"]
  inline_policy {}
}

resource "aws_iam_instance_profile" "testkitchen" {
  name = "test-kitchen"
  role = aws_iam_role.testkitchen.name
}
```

Trying not to use classical remote access creates a problem with Test-Kitchen: It only supports SSH and WinRM for this purpose. But other transports are available as plugins to extend support to other protocols.

On the other hand, similar functionality exists for other tools like InSpec, Chef Target Mode, or `knife bootstrap`: the [Train transport framework](https://github.com/inspec/train). It has a rich ecosystem of standard and more exotic remote access protocols, including AWS SSM.

By combining two tools, we can achieve our goal of SSH/WinRM-less access to our development machines:

1. [`kitchen-transport-train`](https://github.com/tecracer-chef/kitchen-transport-train/) creates the necessary glue to use TK with Train
2. [`train-awsssm`](https://github.com/tecracer-chef/train-awsssm) allows executing commands (RunCommands) via AWS SSM

You can install both with `chef gem install kitchen-transport-train train-awsssm` (if you use Chef Workstation) or `gem install kitchen-transport-train train-awsssm` (if you use plain Ruby).

As this technique uses non-interactive execution, this will sadly make your `kitchen login` commands unusable. But you can use AWS' built-in `aws ssm start-session --target i-123456789012` commands instead.

Our alternative `kitchen.yml` file now looks like this:

```yaml
---
driver:
  name: ec2
  subnet_filter:
    tag: 'Name'
    value: '*private*'
  metadata_options:
    http_tokens: required
    http_put_response_hop_limit: 1
    instance_metadata_tags: enabled
  instance_type: t3a.medium
  iam_profile_name: test-kitchen
  skip_cost_warning: true
  tags:
    CreatedBy: test-kitchen

transport:
  name: train
  backend: awsssm
  execution_timeout: 600 # non-session based, so increase timeout per command

platforms:
 - name: amazon2
 - name: amazon2-arm64
   driver:
     instance_type: t4g.nano
```

We changed our `subnet_filter` to use subnets marked as private and skipped `associate_public_ip` and `interface`.

To enable SSM access, we have to specify the `iam_profile_name` we created with the needed permissions.

Our new `transport` section configures TK to use `train` and `awsssm` for execution. As this is not session-based, but non-interactive, we need to increase the `execution_timeout` to the maximum estimated Chef run duration.

Support for the quicker AWS SSM Session Manager is currently in development. Still, it will need a while as this uses a [custom AWS binary protocol](https://github.com/bertrandmartel/aws-ssm-session/blob/master/src/ssm.js), which is a bit light on the debugging information.

## Summary

Now you know how to use AWS much easier with Test-Kitchen. The solutions should eliminate the need for a personalized `kitchen.local.yml` or even a global configuration. Only tag a subnet, and this will work.

I would be happy about feedback regarding the non-SSH/WinRM alternative, as I wrote both plugins. If you encounter problems or have feature requests, open an issue on the corresponding repositories.

