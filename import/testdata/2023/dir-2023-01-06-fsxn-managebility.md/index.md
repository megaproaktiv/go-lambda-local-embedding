---
author: "Thomas Heinen"
title: "FSx for NetApp ONTAP Manageability Options"
date: 2023-01-06
image: "img/2023/01/pexels-daan-stevens-939331-41-logos.png"
thumbnail: "img/2023/01/pexels-daan-stevens-939331-41-logos.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "fsx", "netapp"]
---
While Amazon FSx for NetApp ONTAP (FSxN) seems relatively easy on the AWS level, it is vastly more powerful if you pick another way to manage it.

This post will look at a quick run-down across the AWS Web Console, NetApp's BlueXP, various APIs, and the CLI.

<!--more-->

As storage is usually regarded to be a "boring" topic among cloud users, the awareness of all the different features and possibilities is highly limited in this group. We are so used to [AWS' concept of snapshots](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSSnapshots.html#how_snapshots_work), that we take this for granted and likely believe it is the only way to do them.

The deeper you look into traditional, on-premises data centers, the more you notice how naive some assumptions are. NetApp has long been able to [instantly clone multi-Terabyte volumes](https://community.netapp.com/t5/Tech-ONTAP-Articles/Back-to-Basics-FlexClone/ta-p/84874) for actively reusing them as writable copies - only needing space for the differences. Try cloning a big EBS volume, and you will immediately notice it is neither fast nor cheap.

Of course, a lot of the functionality in NetApp's ONTAP operating system is not valuable for cloud applications. While classical data centers might need [connections to tape libraries](https://docs.netapp.com/us-en/ontap/ndmp/index.html) or support for non-Ethernet network types ([FibreChannel](https://fibrechannel.org/), [InfiniBand](https://en.wikipedia.org/wiki/InfiniBand)) or protocols like [NVMeOF](https://www.storagereview.com/review/nvme-nvme-of-background-overview) or [GPU Direct Storage](https://community.netapp.com/t5/Tech-ONTAP-Blogs/Optimize-GPU-Accelerated-Workloads-on-NetApp-Storage-Systems-using-NVIDIA/ba-p/432583), the cloud is more straightforward: Usually there is only one way of doing things or a handful of alternatives. There is no need to support every available combination of protocols and hardware, high granularity on performance limits, or minimized latency.

As a consequence, what we usually see as AWS users is only a small subset of the features from ONTAP (which powers FSxN under the hood). But as both AWS and ONTAP are highly API-driven, we can go deeper into the management layers if we have a need for it.

## AWS APIs and UIs

AWS built an abstraction layer over the existing NetApp APIs and added its own infrastructure management under the hood. In the end, when creating your new FSxN filesystem, there will be a need to deploy a pair of virtual ONTAP instances and orchestrate this process. This fact also explains the time needed for this process, and you can look into all those details from the ONTAP CLI (which we will look at later).

So AWS needed to balance different things when creating FSxN:

- ease of use for people without a Ph.D. degree in storage technology
- limitations on the AWS infrastructure layer (latency, technologies, maintainability, cost)
- implementation time for exotic/unavailable features
- interface consistency between all AWS service offerings
- ...

As a result, the AWS Web Console and all AWS tools, like SDKs, CloudFormation, and CLIs, cover the most common use cases.

![AWS FSxN Console](/img/2023/01/fsxn-manageability-awsfsxn.png)

The most basic functionality for FSxN is the management of data volumes and their performance, access from Linux (NFS) or Windows (CIFS), basic backup functionality, monitoring, and high availability.

## BlueXP (formerly: Cloud Manager)

NetApp is also a big software vendor with its own portal to manage products. Previously branded as "Cloud Manager", their [BlueXP](https://www.netapp.com/bluexp/) portal offers a graphically-oriented interface to manage the majority of their storage products. This portal provides a unified view into systems and services cross-account, cross-cloud and cross-data centers. You can quickly get detailed metrics, set up data replications, or manage data across all of them from a unified workspace.

While it is a similarly stripped-down interface as the AWS Web Console, it also includes add-ons: managing Kubernetes Storage, more complex backup/replication strategies, Ransomware and Data Loss Protection, and even Cost Optimization (under their [NetApp Spot](https://spot.io/) brand).

![BlueXP Volume View](/img/2023/01/fsxn-manageability-bluexp.png)

## Native ONTAP Management

But if you have more specific requirements or things that do not fit the usability/roadmap yet, you find references to NetApp's native management functionality inside of [AWS' FSxN documentation](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/managing-resources-ontap-apps.html).

You can find references to "management endpoints" inside the AWS Web Console. These addresses provide access to the cluster-level (FSxN properties, `fsxadmin` user) or the containers that manage storage sections (Storage Virtual Machines provide native multi-tenancy functionality, `vsadmin` user). To use these, you have to open up your FSxN security group either for the SSH-based CLI (TCP port 22) or the HTTPS-based API (TCP port 443). Please notice, that you first have to assign passwords to the users, because they are locked by default.

The modern [REST API](https://library.netapp.com/ecmdocs/ECMLP2884821/html/#/) (`/api`, since ONTAP 9.6) offers a consistent way to query and manage most ONTAP properties. It follows most of the [REST API best practices](https://restfulapi.net/) to provide predictable and understandable methods of handling most tasks the base operating system is capable off.

While usually also deployed on the HTTPS endpoint, NetApp's [System Manager](https://www.netapp.com/data-management/oncommand-system-documentation/) is currently disabled on FSxN. This interface covers management functionality right in between that of BlueXP and the base API/CLIs. It is not yet enabled on FSxN as of January 2023, so the comparison table does not include a column for this tool.

It is also worth mentioning that there has been another API before the modern REST API. The ZAPI was based on XML and offered more extensive feature coverage than its modern cousin. The difference primarily consists of outdated features, deprecated functionality, or available development time.

As this old API is [deprecated by January 2023](https://mysupport.netapp.com/info/communications/ECMLP2880232.html), it is not of interest anymore.

## Infrastructure as Code

Provisioning resources in a reliable and repeatable way has become the norm. Commonly grouped under the term "Infrastructure as Code" (IaC), different tools compete for advanced administrators and DevOps specialists.

In the case of FSxN, the more popular tools are roughly equivalent to the available UIs and APIs. This means that we have CloudFormation/[CDK](https://docs.aws.amazon.com/cdk/v2/guide/home.html), [Terraform](https://www.terraform.io/), [Ansible](https://www.ansible.com), and other tools at our disposal for those tasks.

My preference for FSxN would be to use Terraform for orchestration (creating filesystems and volumes) and Ansible for the more specific configuration (replication, detailed settings, local user management, hardening). While it hurts me a bit personally as a [Chef Infra](https://www.chef.io) expert, it is a fact that NetApp themselves are developing the [Ansible integration](https://github.com/ansible-collections/netapp.ontap) and keep it up to date. If the vendor manages it, there is little benefit in choosing another tool.

## Comparison Table

Now to the central part of this article: the comparison. I also kept this high-level and specifically want to point out the use cases you might have in real-world scenarios. Most importantly, you will have to deviate from AWS' native tooling for large parts of using Windows/CIFS, profit from the instantaneous cloning/replication features, or need block-based iSCSI devices.

| Feature                                           | AWS[^1]     | BlueXP[^2]  | REST API[^3] | CLI[^4] |
| ------------------------------------------------- | ----------- | ----------- | ------------ | ------- |
| AWS native backup                                 | x           | -           | -            | -       |
| Syncing with non-ONTAP (NFS, SFTP, Azure, GDrive) | -           | x           | -            | -       |
| _in-depth and deprecated features_                | -           | -           | -            | x       |
| Create Cluster/FSxN Filesystem                    | x           | x           | -            | -       |
| NFS Volumes                                       | limited[^5] | x           | x            | x       |
| CIFS Volumes                                      | limited[^6] | limited[^6] | x            | x       |
| Volume Capacity Management                        | -           | x           | x            | x       |
| iSCSI Volumes                                     | -           | x           | x            | x       |
| Volume Cloning                                    | -           | x           | x            | x       |
| Native Snapshots                                  | -           | x           | x            | x       |
| Syncing with other ONTAP devices                  | -           | x           | x            | x       |
| Ransomware Protection[^7]                         | -           | x           | x            | x       |
| CIFS share management                             | -           | -           | x            | x       |
| S3 Volumes[^7]                                    | -           | -           | x            | x       |
| Data migrations, manual cluster failovers, etc.   | -           | -           | limited      | x       |

Even for features marked as "supported", REST API and CLI often support various additional options.

[^1]: also valid for CloudFormation/CDK, [`hashicorp/terraform-provider-aws`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs), but not for [`hashicorp/awscc`](https://registry.terraform.io/providers/hashicorp/awscc/latest) (no FSx resources yet)
[^2]: also mostly valid for [`NetApp/terraform-provider-netapp-cloudmanager`](https://registry.terraform.io/providers/NetApp/netapp-cloudmanager/latest/docs)
[^3]: also mostly valid for [Ansible `netapp.ontap`](https://github.com/ansible-collections/netapp.ontap) (managed by NetApp)
[^4]: also mostly valid for ZAPI before and including ONTAP 9.12
[^5]: no selection of NFS protocol or export policies
[^6]: only with Active Directory, not with local users or groups in Workgroup mode. Management via Windows MMC possible.
[^7]: not available on AWS yet

## Summary

While FSxN might seem pretty straightforward at first glance, the possibilities are much more varied, and its administration varies widely depending on your needs. You will need to combine tools to reach your goals, like Terraform+Ansible or CDS+REST API. And in operations for non-trivial systems, access to the CLI interface is required.

That being said, [tecRacer](https://tecracer.de) is both an AWS Premier Tier Services Partner and a NetApp Gold Partner. If you want to use Amazon FSx for NetApp ONTAP for advanced use cases, migrations, or hybrid cloud scenarios, feel free to contact us at [aws-sales@tecracer.de](mailto:aws-sales@tecracer.de)
