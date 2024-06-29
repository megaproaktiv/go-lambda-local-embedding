---
author: "Thomas Heinen, Kobi Biton"
title: "Multi-AZ Block Storage for EKS - Solution Overview"
date: 2022-06-09
image: "img/2022/06/pablo-garcia-saldana-lPQIndZz8Mo-unsplash.png"
thumbnail: "img/2022/06/pablo-garcia-saldana-lPQIndZz8Mo-unsplash.png"
toc: true
draft: false
categories: ["aws"]
tags: ["aws", "eks", "fsx", "netapp"]
---
Did you already encounter an application on EKS which does not perform well with EFS storage or which even needs dedicated block storage with multi-AZ capabilities? In this case, we have prepared something for you:

We now support you with creating multi-AZ block storage for your EKS clusters and add facilities to monitor, manage and troubleshoot them with easy building blocks.

<!--more-->

Kubernetes has become the de-facto standard in container management and orchestration.

At tecRacer, we specialize in supporting customers such as Financial Services, Logistics, Engineering, and more on how to architect for Amazon Web Services. We have transformed many customer workloads to work on Amazon EKS and EC2-based Kubernetes clusters.

Those projects clearly show one thing: Kubernetes has challenges, mainly for stateful applications. In an ideal world, every application would follow the [12-factor principles](https://12factor.net/), which work perfectly on Kubernetes, but the reality is way different. Many applications out there still heavily rely on state and require atomic storage performance, among them: databases, distributed cache systems, and the list goes on and on.

It is worth mentioning that customers often use fully managed shared file system implementations such as Amazon EFS for this challenge. These can help Kubernetes applications be highly available because all data written to EFS gets replicated to multiple AWS Availability Zones.

Running those applications in Kubernetes usually means that the operator needs to set up zonal cloud architectures and introduce scheduling constraints. This path puts a lot of burdens and heavy lifting on the cluster management and application lifecycle.

The operator needs to create a Managed Node Group per availability zone per stateful application on AWS, which mandates localized block storage devices such as Amazon EBS. They will also need to ensure that workloads get scheduled in the proper availability zone to access their storage.

One more challenge is synchronizing data between AZs, which is yet to be introduced as part of Amazon EBS. As a result, customers often need to implement and maintain a custom mechanism to achieve that [^1].

## The Solution

Our solution for this problem is a combination of Amazon FSx for NetApp ONTAP (short: FSxN), NetApp Trident as CSI (Container Storage Interface) driver, and additional components for operations which we customize to your needs.

![Architecture](/img/2022/06/multi-az-block-storage-eks-simple.png)

Amazon FSx for NetApp ONTAP offers a sophisticated third-party storage solution as managed service. It can offer both file and block storage synchronously replicated between two AZs (Multi-AZ deployment). Any failover of storage backends is designed to occur transparently and non-disruptively via ISCSI multipathing, and management of Kubernetes storage is handled automatically by Trident without needing application modification.

The base technologies have shaped classical data center designs for a decade and now are available in cloud solutions like Amazon EKS with our integration. Failover and failback of storage between AZs usually should not result in Pod restarts, only slight latencies while the backend connections switch over on the OS level.

It is always recommended to test this failover on a regular basis and in parallel keep and monitor your backups processes regardless of this solution. Your tests should let you know if your application can handle a crash-consistent start. 

This combination results in architectures that need only one EKS Managed Node Group per application and can use a well-maintained managed service to provide block storage. Other features of FSxN are sub-millisecond latency, 100,000s of IOPS, and more efficient storage thanks to deduplication, compression, and compaction.

For Day 2 operations, we also provide a Grafana-based storage metrics monitoring solution (based on NetApp Harvest), automated configuration compliance checks, authenticated access to the base ONTAP APIs for existing third-party tooling, Cloudwatch Logs integration, and various Systems Manager documents to perform routine tasks on the FSx/ONTAP storage backend.

## Block Storage: Amazon FSx for NetApp ONTAP

In September 2021, FSxN launched as a native AWS service. The underlying ONTAP system has been proven in enterprise data centers for over 15 years - for geo-redundant replication of storage, block-level replication, and automatic failover of storage.

While setting up NetApp hardware or their virtualized counterparts needs expertise in configuring replication relationships and failover mechanics, FSx for NetApp ONTAP provides an easy-to-use UI with all benefits of AWS APIs. Deploying a Multi-AZ storage system, modifying its parameters, and creating network storage for your systems is easy and does not need much experience.

You can set up a filesystem in two different ways:

- Multi-AZ to allow the failure of one AZ without impact on storage
- Single-AZ for a price-optimized deployment which still includes redundancy

While filesystem (NAS) functionality is visible in the AWS console, its block-based (SAN) counterpart is not. It needs knowledge of data center technologies like ISCSI, Multipathing, LUN masking, and specific ONTAP CLI commands - which is not a given in a cloud-first environment.

## Kubernetes CSI: NetApp Trident

Connecting storage to Kubernetes can be a challenge, especially if it is storage that is not local to the nodes. The plugin architecture of Kubernetes includes Container Storage Interfaces (CSIs), which manage connections to different storage backends and still provide native Kubernetes storage classes to your applications.

NetApp Trident is the CSI driver for all products by NetApp - not only FSxN but also E-Series, HCI and SolidFire. It is installed via Helm charts, the legacy Trident installer, or directly as Custom Resource Definitions (CRDs) via `kubectl`. You can configure access to backend storage via NAS (based on NFS) and SAN (ISCSI).

Trident deploys an operator to manage the whole architecture, a DemonSet on all cluster nodes to manage CSI operations, and a centralized orchestrator to keep track of volume requests and assignments. Storage backends get configured and then assigned to Kubernetes' native storage classes. The CSI driver handles all management of Physical Volumes, automatically creating and destroying volumes.

As a result, applications do not need to be adjusted to take advantage of all Trident capabilities - they only use the predefined storage classes and can run unchanged. You specify their storage class, and the management of any persistent volumes will be automatic.

## Operations

### Metrics Monitoring

Based on the industry-standard Grafana for metrics monitoring, storage data is collected and displayed in specialized dashboards. On the one hand, a Harvest collector constantly queries the underlying ONTAP API for metrics about latencies, bandwidth, IOPS, and potential performance bottlenecks. In addition, metrics provided by the Trident CSI driver add a perspective to Kubernetes-specific metrics on the storage level.

While Harvest is available as a standalone product, integrating it with FSxN and adding the Trident-generated metrics to dashboards benefits seeing potential problems arising and offers insights into storage savings and optimization possibilities.

### ONTAP Audit Logging

Currently, FSxN has no integration with any AWS logging solution. As a result, there is limited visibility into underlying problems or any activities on the management level. To keep track of changes in topology and configuration, we deploy a bridge solution that allows the base cluster to send auditing events to a dedicated CloudWatch Logs stream. Consequently, you can use all AWS tooling like log metrics, alerting, and eventing.

### Configuration Compliance

Visibility into infrastructure configuration issues is critical for operations. We integrate regular checks of the base ISCSI and multipathing configuration into Systems Manager Configuration Compliance dashboards. This way, you will get notice of accidental misconfigurations and runtime conditions.

### Third Party Tool Integration

While many basic tasks are possible in the AWS FSx console, troubleshooting and optimization still need access to the underlying ONTAP system. FSxN does provide a native REST API for this, but this is not readily available outside of the VPC.

By deploying a specialized API gateway solution, our architecture allows access to these APIs with third-party tools or custom tooling like  Lambda functions. No need to use SSH from within Jumphosts in your VPC. You can connect to the deployed API gateway securely.

### Pre-Packaged ONTAP SSM Documents

In different situations, even when initially setting up logging to CloudWatch, there is a need to issue commands to the internal REST API. With a subcomponent of the solution, this is easy - a Lambda function connects to the API endpoint and can replay commands deployed as regular SSM automation documents.

These documents create a library of reusable ONTAP configuration snippets for logging, performance tuning, diagnostics, etc.

## Engagement Timeline

This solution is a consulting offering, which means it is not a fixed price offering but results in a design to supply you with an optimized solution for your specific use cases.

If you are interested in this solution, we will start with a kick-off meeting to determine your needs and existing infrastructure.

We will create a Proof-of-Concept cluster with all technologies in the first phase. You can then try the different performance and failover characteristics you need to determine the feasibility in your environment. Everything is code-driven and reproducible, so it is easy to tear down and rebuild the infrastructure with little effort (e.g., A/B tests).

After the PoC, we will work on the pillars of a production-ready infrastructure for your applications. All steps rely on the AWS Well-Architected practices, practical Kubernetes knowledge, specific applications and networking connections, and integrations with your existing tooling. We can also create this in Terraform and YAML for reusability, depending on your cloud strategy.

For your operations, we then deploy the listed components for monitoring, logging and access depending on your needs.

## Get in Touch

You can find our offer on the [AWS Marketplace](https://aws.amazon.com/marketplace/pp/prodview-wllssiaxvpv3m), for more information contact us now at [aws-sales@tecracer.de](mailto:aws-sales@tecracer.de). We will schedule a call with one of our specialists to answer your questions.

[^1]: You can use open source software for replication data like [DRBD](https://linbit.com/drbd/)
