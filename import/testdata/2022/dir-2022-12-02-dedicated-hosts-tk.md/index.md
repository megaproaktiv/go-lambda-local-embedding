---
author: "Thomas Heinen"
title: "Dedicated Hosts with Test Kitchen"
date: 2022-12-02
image: "img/2022/12/pexels-brett-sayles-2881233-1920-4-1.png"
thumbnail: "img/2022/12/pexels-brett-sayles-2881233-1920-4-1.png"
toc: false
draft: false
tags:
  - devops
categories:
  - AWS
  - Chef
  - Test Kitchen
---
Sometimes, you need to deploy software for tests with special licensing terms. To solve this, AWS offers Dedicated Instances and Dedicated Hosts - and now you can use them with [Test Kitchen 3.14](https://kitchen.ci) in your developer workflows.

<!--more-->

## Why Dedicated Hosts

While there are several reasons to use dedicated hosts, most are related to [licensing terms](https://aws.amazon.com/ec2/dedicated-hosts/). Some software vendors only issue licenses to be used on single-tenant systems. Others even require you to license all hosts where their software could run, not only those which do.

As a Dedicated Host (DH) is only available for your company and will never run instances from other customers, this fills those requirements. Licensing can be touchy, so please ask your license manager or software vendor if you have conditions like these.

Another reason is security-related: Theoretically, there could be bugs in the base hypervisor, which allow attackers to break out of an instance and then access other machines on the same hardware. While these cases are scarce, very sensitive data might be required to be run on dedicated machines.

## Why not Dedicated Instances

Dedicated Instances can solve this issue. Their use cases are very similar to Dedicated Hosts while being much cheaper, but they might not fulfill vendor-specific licensing requirements.

It is easy to use them with Test Kitchen, as it is just a matter of defining `tenancy: dedicated` in your `kitchen.yml`, and it will work out of the box. Please look into the AWS pricing pages to get more information on the additional costs.

This setting has been available since [version 1.3.0 of Test Kitchen](https://github.com/test-kitchen/kitchen-ec2/blob/v1.3.0/CHANGELOG.md)

## Dedicated Hosts Basics

In contrast to regular or Dedicated Instances, using Dedicated Hosts will require an additional step: Allocation of a host. This is a separate action in the web console or CLI/SDK, which needs `ec2:AllocateHosts` privileges.

You decide which type of instances you want (a DH will only support one specific family), which Availability Zone to pick, and if it is available for automatic use with corresponding instances.

![EC2 Dedicated Hosts](/img/2022/12/dedicatedhosts-ec2.png#center)

For our purposes in Test Kitchen, please remember to select auto-placement as it is not yet possible to address a specific task in the driver.

After the host gets provisioned within seconds, you can use it to place new instances on it until it is at capacity. At that point, you either have to stop other instances or need more hosts.

Remember to deallocate unnecessary hosts quickly to avoid paying for unused capacity. And notice that `t3` Dedicated Hosts are much more expensive in this case[^1].

### Lifecycle

As you can see, the additional lifecycle management of DHs can be challenging.

You can manually allocate and deallocate hosts but risk having unused ones running too long without being noticed.

The `kitchen-ec2` driver can manage this overhead if you enable it. Both the `allocate_dedicated_hosts` and `deallocate_dedicated_hosts` settings are available to make granular changes, depending on if you work locally (set both to `true`), if you use this in a CI/CD system (likely allocate, but do deallocation on a schedule), or if you have some exceptional circumstances (like `mac` hosts)

On the AWS side, a little hidden gem is under the "AWS License Manager" service, where you can use the built-in Host Resource Groups feature.

![License Manager](/img/2022/12/dedicatedhosts-licensemgr.png#center)

This feature automatically allocates new hosts if additional capacity is needed or deallocates empty hosts.

But there is a caveat: It does not support all instance types but only a subset.

## Using it with Test Kitchen

To use this, you need to update the `driver` properties in your `kitchen.yml` as a top-level setting or as a more specific configuration under `platforms`.

The first property is `tenancy`, which usually is `default` - so it will allocate on-demand instances. You can set this to `dedicated`, which means Dedicated Instances, or to `host`.

If you use `host`, you should either use one of the external lifecycle management systems described above, or you need to set the `allocate_dedicated_hosts` setting to `true`.

Likewise, you might want to use the deallocation feature via `deallocate_dedicated_hosts` as `true`.

Notice that you need to specify the `availability-zone` attribute, or the AWS API will not know where to place the new host.

Finally, the following is a snippet for a working Dedicated Hosts configuration with Test Kitchen:

```yaml
driver:
  allocate_dedicated_host: true
  deallocate_dedicated_host: true

instances:
  - name: ubuntu-20.04
    driver:
      availability-zone: eu-west-1
      tenancy: host
```

[^1]: See the [AWS pricing page for dedicated hosts](https://aws.amazon.com/ec2/dedicated-hosts/pricing/#On-Demand_Pricing)
