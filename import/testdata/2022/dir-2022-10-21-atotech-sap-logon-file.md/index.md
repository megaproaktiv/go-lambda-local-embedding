---
title: "Deploy a cloud-native SAP Landscape File using S3 PrivateLink"
author: "Leon Bittner"
date: 2022-10-21
toc: false
draft: true
image: "img/2022/10/202297_AWS_Cloud_Migration_Diagram (2).b3473264eab040033ff44cbf9a5743a209c30563.png"
thumbnail: "img/2022/10/202297_AWS_Cloud_Migration_Diagram (2).b3473264eab040033ff44cbf9a5743a209c30563.png"
categories: ["aws", "ads"]
tags: 
    - level-300
    - docker
    - vmware
    - application discovery service
    - ads
    - migration hub
    - proxy
keywords:
    - docker
    - vmware
    - application discovery service
    - ads
    - migration hub
    - proxy
    
    
---

One simple example of transforming SAP solutions with the help of cloud-native technologies is the SAP Landscape File. The SAP Logon File is a global repository of each system that an SAP landscape consists of. It needs to be reachable from every client accessing SAP systems.  This article describes, how we deployed a serverless SAP Landscape File solution at MKS Atotech.

<!--more-->

## About MKS Atotech

Atotech is a brand of MKS, a leading specialist in foundational technologies for semiconductor manufacturing, advanced electronics, and specialty industrial applications. At MKS, Atotech delivers chemistry, equipment, software, and services to a variety of markets for example consumer electronics, communication infrastructure, automotive, and much more. MKS Atotech is a global player with more than 4.000 employees, 15 technology centers, and 17 production sites around the world. With the help of tecRacer and other cloud technology partners, MKS Atotech is on a journey to migrate all IT workloads to the AWS cloud.

## Challenge 

At the moment, many companies rely on old-fashioned file servers for the hosting of SAP landscape files. However, migrating this to a VM running in the cloud would be a bad choice. Cloud providers like AWS are able to allocate IT resources extremely efficiently. But in order to benefit from this, the end customer must also choose the right services to operate their solution. In this case, we have chosen Amazon S3 (Simple Storage Service) as the core of our solution. 

S3 is a crucial service for every cloud project. Not only does it provide high availability and durability with every file being spread over three different datacenters. It is also much more cost-efficient than running servers. At the moment of writing, one Gigabyte costs 0,023 USD per month. It is important to note though, that S3 works different than regular file storage. S3 is an object store. The only way to access files is via http(s) requests. S3 also works different than on-premise IT components from a security perspective. To prevent unauthorized access you work with the IAM and bucket policies instead of firewalls. Of course, public S3 Buckets are a big no-no because they often contain sensitive data.
  
## Solution

Our goal is to host the SAP landscape file in S3 and make it available only to the Atotech network. tecRacer has built a sophisticated hybrid-cloud architecture at Atotech, which supports us in building our solution. All traffic between the Atotech data centers and the AWS Cloud takes place via a so-called Direct Connect. This is a dedicated, encrypted connection between the on-premises and AWS. On the AWS-side, the Direct Connect is attached to a Transit Gateway, which connects multiple VPCs as well as the on-premise network to each other. 

![Architecture for internal-facing S3 Bucket](/img/2022/10/internal-bucket-architecture.png)

To get the traffic flowing to our S3 bucket, we now need a way to directly route it through the Direct Connect. This is why we deployed an Interface S3 Endpoint in a VPC that is attached to the Transit Gateway. Simply speaking, this endpoint consists of three virtual network cards (one for each availability zone) that have their own private IP adress. Those endpoints are quite handy because they allow private connectivity to AWS services such as S3, API Gateway or Systems Manager without the need of routing through the internet. Because of the hybrid infrastructure in combination with IP routing, we can reach the S3 Endpoint from the on-premise network.

Now that we are set up from a networking perspective, we still need to restrict the access to our Bucket to only make it accessible from the Atotech intranet. To achieve this, we create a Bucket with a policy that allows only traffic that originates from Atotech’s VPC endpoints I just described. 

```json
 {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "s3:ListBucket",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::xxxx/*",
                "arn:aws:s3:::xxxx"
            ],
            "Condition": {
                "StringEquals": {
                    "aws:SourceVpce": [
                        "vpce-xxxx"
                    ]
                }
            }
        }
    ]
}
```

If traffic would originate from any other source than the specified VPC endpoint, for example, the Internet, it would get blocked by the bucket policy. With this policy in place, we cannot use the public S3 endpoint URL anymore: [bucket-name].s3.[region].amazonaws.com/objectkey. Instead, we use the DNS name of the S3 endpoint: vpce-xxxxxxxxxxxxxx-xxxxxxxx.s3.region.vpce.amazonaws.com/bucket-name/objectkey.

But since this URL is not very handy, we can create a custom CNAME entry that points to the endpoint. For example: mybucket.mycompany.local. Please note, that in this case the Bucket name has to be the same as the CNAME entry. We can check if everything works by trying to access the SAP Logon File via the browser or by doing an nslookup (URLs and IP adresses are pseudonymised):

![Checking if the routing works correctly with nslookup](/img/2022/10/internal-bucket-nslookup.png)

We can see that the internal DNS name resolves to the IP adresses of the three VPC endpoints. Now, the last thing we need to do is deploying the URL of the S3 bucket to the SAP GUI clients of the company. The most common way to do this is via GPO.

## Conclusion

This article highlighted how to deploy an internal-facing S3 Bucket that serves the SAP Logon File to all the clients accessing SAP Systems. The solution relies on hybrid cloud connectivity and bucket policies to prevent external access. By using S3 as a serverless technology, it is a highly available and cost-effective alternative to running a file server.
