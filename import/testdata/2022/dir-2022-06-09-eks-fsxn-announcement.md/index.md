---
author: "Thomas Heinen"
title: "Multi-AZ Block Storage for EKS - New Consulting Offer"
date: 2022-06-09
image: "img/2022/06/pablo-garcia-saldana-lPQIndZz8Mo-unsplash.png"
thumbnail: "img/2022/06/pablo-garcia-saldana-lPQIndZz8Mo-unsplash.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "eks", "fsx", "netapp"]
---
Today we announce a new consulting offer on the AWS Marketplace: Multi-AZ Block Storage for EKS. Finally, you can use block-based storage for your specialized workloads, which is highly available.

<!--more-->

You can now contact us about our new solution to provide highly available block-based storage for your EKS clusters. With this new architecture, you can finally run applications that highly depend on atomic storage performance (such as databases or distributed cache systems) in a multi-AZ scenario.

Classical EKS architectures rely on network-based file services like EFS or local block storage like EBS. While the first might not be optimal due to latency or file metadata overheads, the second is not capable of multi-AZ.

Solving these issues often adds overhead on EKS architecture with zonal architecture, scheduling constraints, or even self-managed replication solutions.

With our consulting offer, you first get a Proof-of-Concept cluster to experiment with and test against your business objectives. Our consultants will then look at your workload and requirements to adapt everything for your needs in a second step.

As every architecture is only as good as its maintainability, we also add components for:

- using Amazon FSx for NetApp ONTAP with third-party storage tools (via ONTAP REST API)
- investigate metrics on internal latencies and limitations
- integrating with CloudWatch Logs for storage backend auditing
- continuous configuration compliance checks.

Read more on the [AWS Marketplace](https://aws.amazon.com/marketplace/pp/prodview-wllssiaxvpv3m), our [detailed solution blog](https://www.aws-blog.de/2022/06/multi-az-block-storage-for-eks-solution-overview.html), or contact us via [aws-sales@tecracer.de] for more details and offers.
