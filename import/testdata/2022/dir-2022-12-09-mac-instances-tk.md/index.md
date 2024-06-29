---
author: "Thomas Heinen"
title: "Using AWS mac1/mac2 Instances with Test Kitchen"
date: 2022-12-09
image: "img/2022/12/pexels-wendy-wei-1656670-4-1.jpg"
thumbnail: "img/2022/12/pexels-wendy-wei-1656670-4-1.jpg"
toc: false
draft: false
tags:
  - devops
categories:
  - AWS
  - Chef
  - Test Kitchen
---

Everybody who had to write software or work with configuration management for Apple knows of the problems to get access to test machines. AWS does offer both Intel- and M1-based Mac instances now and with `kitchen-ec2` v3.15.0 it is finally possible to use them in your existing workflow.

<!--more-->

AWS announces the availability of Intel-based `mac1` instances in [November 2020](https://aws.amazon.com/about-aws/whats-new/2020/11/announcing-amazon-ec2-mac-instances-for-macos/) and `mac2` M1 instances in [July 2022](https://aws.amazon.com/blogs/aws/new-amazon-ec2-m1-mac-instances/). These are based on Mac Minis in custom rack mounts, integrated into the AWS ecosystem

![Mac in AWS Rack](/img/2022/12/mac-instances-hardware.png#center)

Even in the AWS cloud, those are different: You can only use `mac1`/`mac2` on dedicated hosts (see my [earlier post on kitchen-ec2 3.14]()) and they have a minimum 24-hour allocation and billing period. Within this time, any try to deallocate a `mac1`/`mac2` host will fail with an error message.

In consequence, you should implement some sort of external lifecycle management to remove unused hosts automatically - either using AWS License Manager Host Resource Groups or using some custom tooling.

![Automated release of Dedicated Hosts](/img/2022/12/mac-instances-hrg.png#center)


## Configuration: Intel-based `mac1`

With `kitchen-ec2` v3.15.0 it is pretty easy to use Apple-based instances.

The older Intel-based `mac1` instances are more expensive - in my preferred region `eu-west-1` they typically are around $25 per day (remember the 24-hour minimum billing here again).

They start fairly quickly and within about 5 minutes you can SSH into them or start your usual `kitchen converge` run.

![Provisioning mac1 with Test Kitchen](/img/2022/12/mac-instances-mac1.png#center)

At this point in time, you can develop and check your implementations. But there is a huge additional caveat: You might already be used to the short-lived instances of Test Kitchen, running rapid `kitchen create`/`kitchen destroy` actions.

This won't work on any AWS-based Apple instance.

As [AWS details in its documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-mac-instances.html#mac-instance-stop), any stop/termination of an instance on `mac1`/`mac2` hosts will automatically initiate the scrubbing workflow, which removes user data so new instances on the same hardware cannot access it. For Intel-based instances, this is documented to take __roughly an hour__. If you are particularly unlucky, this might even be longer if updated firmware is deployed automatically. For more info on the dedicated hosts lifecycle have a look at the [detailed blog on the dedicated host lifecycle by AWS](https://aws.amazon.com/blogs/compute/understanding-the-lifecycle-of-amazon-ec2-dedicated-hosts/).

For this duration, your Dedicated Host will stay in state `pending` and not be usable for your next tests.

Apart from this, adding a `mac1` platform to your Test Kitchen configuration is straightforward:

```yaml
platforms:
 - name: macos-12.5
   driver:
     instance_type: mac1.metal
     availability_zone: eu-west-1a
     tenancy: host
```

Notice, that the AZ needs to match one of any preallocated dedicated hosts (in that case, do not forget to add the `ManagedBy` tag with `Test Kitchen` as its value).

The `kitchen-ec2` driver will automatically search suitable official AMIs by the given version string (in this case, the AMI name will be searched with `amz-ec2-macos-12.5*`).

If you do not specify anything additionally, Intel-based AMIs will be the default.

## Configuration: M1-based `mac2`

The newer variant of Mac instances is Apple M1-powered and cheaper. In my case, I pay around $17 per day in `eu-west-1`.

But for some reason, using `mac2` is much slower: Their start needs about 25 minutes and the scrubbing workflow is documented to be around two full hours. This duration is highly uncomfortable and limits your workflow despite the new Test Kitchen capabilities.

![Waiting for mac2 scrubbing](/img/2022/12/mac-instances-console.png#center)

```yaml
platforms:
 - name: macos-12.6-arm64
   driver:
     instance_type: mac2.metal
     availability_zone: eu-west-1a
     tenancy: host
```

It is particularly important to specify the architecture in the platform name (`-arm64`) which will influence the AMI search pattern to use the `arm64_mac' architecture used by AWS internally.

## Further Steps

After you created your instance, you can work with it normally. Be it by using `kitchen login`, classical SSH, or switching over to SSM if you attached an IAM profile with the proper privileges.

![VNC Access](/img/2022/12/mac-instances-vnc.png#center)

The `aws-samples` GitHub repository contains a lot of information on what you can do next. You can read on their [Mac Getting Started - Step 3](https://github.com/aws-samples/amazon-ec2-mac-getting-started/blob/main/steps/03_connect_and_enable.md) page how to enable VNC access, resize the virtual display or the root volume, etc.

## Summary

While it is now possible to use `mac1`/`mac2` with Test Kitchen, cost and start/stop durations make this barely usable. You are especially endangered to stack up massive bills if you forget to deallocate your unused Dedicated Hosts.

For the sake of sped-up delivery, let us hope that AWS finds a way to massively reduce the wait times on instance-level actions.
