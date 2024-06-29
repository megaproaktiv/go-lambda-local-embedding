---
author: "Thomas Heinen"
title: "Out-of-Band Bootstrapping with Chef on AWS Systems Manager"
date: 2023-08-01
image: "img/2023/08/markus-spiske-uBcgQA7fwEA-unsplash-41.jpg"
thumbnail: "img/2023/08/markus-spiske-uBcgQA7fwEA-unsplash-41.jpg"
toc: false
draft: false
categories: ["chef", "aws"]
tags: ["aws", "chef"]
---
A modern architecture avoids opening any SSH or WinRM/RDP ports to minimize the attack surface of your systems. Instead, management connections like the AWS SSM Agent should be implemented. But some tools, especially in the configuration management sector, still rely on direct access.

Chef Infra is on track to break this limitation with its new support for out-of-band (OoB) bootstrapping using Knife and arbitrary Train transports.

<!--more-->

{{% notice info %}}
The mechanics in this post depend on two currently approved but not yet merged Pull requests ([Chef PR #13534](https://github.com/chef/chef/pull/13534), [Train PR #742](https://github.com/inspec/train/pull/742))
{{% /notice %}}

## Bootstrapping

"Bootstrapping" commonly relates to enrolling a system into a centralized management system. With Chef Infra, this means installing the Chef Infra agent on the system, preconfiguring it, getting certificate-based authentication set up, and doing a first run with instructions from the Chef Infra Server.

Previously, this was done via SSH or WinRM only, limiting this to systems with direct reachability. The alternative is a manual process where administrators upload the agent to the machine and set up everything by hand.

In this blog, I refer to indirect access as "out-of-band," regardless of the connection protocol. We will use the example of an SSM-managed EC2-node to clarify the new possibilities.

## Knife

As Chef leans heavily on metaphorical names for its tool, `knife` is the standard tool to interact with a Chef Server. It can execute many tasks, from inventory gathering to workflow management and configuration. You can even extend it with custom plugins if you need additional functionality or commands specific to your company or project. To use `knife`, your workstation needs to be known and authenticated to the Chef Server.

It also is responsible for the bootstrapping of new nodes:

```shell
knife bootstrap 203.0.113.17 --node-name webserver-01 --user ec2-user --sudo --bootstrap-version 18.2.7
```

This command will
- determine the connection protocol: SSH (`22/tcp`) or WinRM (`5985/tcp` or `5986/tcp`)
- connect to the machine
- check the operating system
- download the Chef Infra agent (in this case with a fixed version of `18.2.7`)
- ask the Chef Server for a new client identity and certificate
- check for instructions to run (classical Chef run list or assigned policy)

So after this one-line command, you have the node inside your Chef infrastructure and can manage it along with 10,000s or even 100,000s of other nodes centrally. You can read more about the [internals of Chef Bootstrapping](https://docs.chef.io/install_bootstrap/) on the documentation pages.

## General Knife Train Support

Until now, `knife` had hardcoded SSH and WinRM protocols but already used a modular framework called Train under the hood. This framework specifies an abstract interface for command execution, file transfers, and file operations. The interface then gets implemented by different Train Transports, such as `train-winrm`.

With the initially mentioned pull requests, this restriction will be lifted. While API-based transports (like `train-aws` to communicate with the AWS meta-structure) have been turned off for obvious reasons, any installed Train Transport for command execution can now be used. This change enables other protocols like `train-telnet` and out-of-band functionalities like `train-awsssm`.

## Train AWSSSM

For quite a while, an AWS-specific Train Transport has been available under `train-awsssm`. Its primary use so far was for another project called [InSpec, which does compliance checks](https://www.tecracer.com/blog/2020/10/air-gapped-compliance-scans-with-inspec.html) of cloud platforms or operating systems. It is used daily to check 100,000s of EC2 instances in regulated environments with limited connectivity.

Internally, `train-awsssm` uses SSM Run Documents to encapsulate commands which otherwise would be sent over SSH. The SSM agent on EC2 instances will poll for instructions via HTTPS, get these documents, execute them, and return output and exit codes. While this approach is not ideal from a latency perspective, it does include an audit trail of any commands sent via this out-of-band connection.

Development of the more advanced Session Manager connectivity is ongoing in `train-awsssm` but is a considerable amount of work due to the WebSocket-based proprietary binary protocol involved. This added capability will speed up out-of-band access, making the wait worth it.

## EC2 Prerequisites

To enable EC2 instances for SSM, you have to associate them with an EC2 Instance Profile, which has the appropriate privileges. A quick and secure way to ensure this is to use the `SSMManagedInstanceCore` policy. Alternatively, you can enable SSM access on a per-region basis using [Default Host Management Configuration (DHMC)](https://docs.aws.amazon.com/systems-manager/latest/userguide/managed-instances-default-host-management.html).

If you associate an IAM Profile, you must also enable the Instance Metadata Service (IMDS); this feature is accessible via local IP `169.254.169.254`, and preferably set to require version 2. The older version 1 is still supported but has an inherent risk of leaking credentials if your instance has remote file inclusion (RFI) vulnerabilities.

{{% notice note %}}
If you want to use EC2 tags inside your Chef cookbooks, take care also to enable passing them into IMDS (which is not default). Chef Infra will then provide them in the `node['ec2']['tags_instance_*']` attributes.
{{% /notice %}}

To communicate with the Chef Infra server during and after bootstrapping, enable outgoing HTTPS (`443/tcp`) traffic to your server address.

## Workstation Prerequisites

Your local administrator workstation will need to have an updated Chef Workstation installed, which includes the updates to `knife` and `train`. Also, you need to [connect it to your Chef Infra server](https://docs.chef.io/workstation/knife_setup/)

Then, install your out-of-band Train Transport:
- for AWS SSM:
  `chef gem install train-awsssm`
- for VMware Guest Operations Management:
  `chef gem install train-vsphere-gom`
- if you intend to use a console connection:
  `chef gem install train-serial`

## Using OoB Bootstrapping with AWS SSM

Of course, you need to assume an AWS profile that allows you access to the account the EC2 instance in question is in. For managing your AWS credentials, the recommendations are either the traditional [Awsume](https://github.com/trek10inc/awsume/) command or [Leapp](https://www.leapp.cloud/).

Then, you will need the instance's IP address or ID to bootstrap. `train-awsssm` will automatically discover the instance with this information but not try to use any IP connectivity.

While `knife` has a legacy option to specify the connection protocol (`--connection-protocol` or `-o`), you can use Train's under-documented URL notation instead.

This notation will use the Train Transport name as the URL's schema, and you can even add parameters if the Transport offers them:

```shell
knife bootstrap awsssm://i-1234567890/ --node-name webserver-01 --bootstrap-version 18.2.7 --user ec2-user --sudo

# Windows needs an extended timeout for execution
knife bootstrap awsssm://i-1234567890/?execution_timeout=600 --node-name winserver1 --bootstrap-version 18.2.7
```

You can check those additional parameters on the respective GitHub repositories (like [the parameters for `train-awsssm`](https://github.com/tecracer-chef/train-awsssm/blob/v0.3.1/lib/train-awsssm/transport.rb#L9-L14)).

## Future Developments

It is possible to extend the out-of-band capabilities of Train to other Cloud Providers or Hypervisors - the plugin ecosystem makes this very easy.

Other `knife` functionality still relies on the non-Train connectivity options. Hence, an extension to other subcommands or unification (`knife ssh` vs `knife winrm`) is a logical next step.

## List of related Pull Requests

* [Allow more train plugins for the knife bootstrap command · Pull Request #13534 · chef](https://github.com/chef/chef/pull/13534)
* [Fix upload to support individual files · Pull Request #742 · train](https://github.com/inspec/train/pull/742)
* [Make option parsing more robust to avoid type errors · Pull Request #9 · train-awsssm](https://github.com/tecracer-chef/train-awsssm/pull/9)
