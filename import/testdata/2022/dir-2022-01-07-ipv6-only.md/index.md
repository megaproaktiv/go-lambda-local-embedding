---
author: "Thomas Heinen"
title: "One Step Closer to IPv6"
date: 2022-01-07
image: "img/2022/01/ipv6-680x363.jpg"
thumbnail: "img/2022/01/ipv6-680x363.jpg"
toc: false
draft: false
tags:
  - Network
  - Aws
  - VPC
categories:
  - AWS
---
Over many years, we have now read warnings about the exhaustion of available IPv4 addresses. So far, there still seem to be ways and ideas on how to extend their lifetime (by approaching large organizations, using NAT, re-dedication of 240.0.0.0/4, and so on).

Switching over to the much-dreaded IPv6 sounds easy, but even minor things can still cause problems.

So what is the current state of AWS with this topic? And how did the landscape change during re:Invent 2021?

<!--more-->

## IPv6-only Subnets

There were multiple announcements related to IPv6 during re:Invent 2021, most notably the availability of IPv6-only Subnets, which got an in-depth [AWS Blog Post](https://aws.amazon.com/blogs/networking-and-content-delivery/introducing-ipv6-only-subnets-and-ec2-instances/).

Still, if you read the article carefully, you will find one gap: If you create an IPv6-only subnet, how would you communicate with non-IPv6 targets?

Luckily on the very next day, a feature for this was released but lacked an in-depth blog post: [DNS64/NAT64 Support](https://aws.amazon.com/about-aws/whats-new/2021/11/aws-nat64-dns64-communication-ipv6-ipv4-services/)

## DNS64 and NAT64

The two addressing schemes, IPv4 and IPv6, are fundamentally different in syntax, so bridging the gap seems impossible. But as IPv6 is much larger, it is actually no problem to reserve an area inside of it for "wrapping" old-style addresses.

This wrapping is the idea behind NAT64[^1]: It uses the IPv6 Well Known Prefix `64:ff9b::/96`[^2] which has a size of 32 bits, exactly the addressing range of IPv4, and maps the legacy addresses onto it:

| IPv4         | IPv4 octets in Hex | IPv6              |
| ------------ | ------------------ | ----------------- |
| 1.1.1.1      | 01 01 01 01        | 64:ff9b:0101:0101 |
| 72.21.210.29 | 48 15 D2 1D        | 64:ff9b:4815:d21d |

So if you have a network device that is responsible for handling the `64:ff9b::/96` prefix, it can translate incoming IPv6 requests into an IPv4 address, send them out and translate the response back into IPv6. While the implementation is much more complex, NAT64 generally provides this exact mechanism.

But how would our IPv6-only clients even come up with those IPv6-wrapped addresses, if it cannot handle them natively? That is the second part of the equation: DNS64[^3]. If you query a name server with DNS64 support, it will either return such a wrapped address (if there is only an IPv4/A-type record for the target) or the native IPv6 address (if there is an IPv6/AAAA-type record available).

![Translated IPv4 address](/img/2022/01/ipv6-translated-ipv4.png#center)

As you can see, the IPv4 address gets translated into its IPv6 representation automatically by the DNS64-enabled and VPC-integrated nameserver.

So there are three components needed for DNS64/NAT64 on AWS:

1. Activating this feature on the IPv6-only subnet to adjust DNS lookups
2. A NAT Gateway which is in a subnet with both IPv4/IPv6 addresses
3. The proper route table entry to hand requests to `64:ff9b::/96` over to the NAT Gateway

![Activate DNS64 on subnet settings](/img/2022/01/ipv6-subnet-dns64.png#center)

## Route Tables

With these alternatives, we end up with four different routing table designs. The only type that does not have any IPv4 capabilities is the "Public IPv6-only" subnet type.

*Public Dualstack*

| Prefix    | Next Hop | Remark      |
| --------- | -------- | ----------- |
| 0.0.0.0/0 | IGW      | Public IPv4 |
| ::/0      | IGW      | Public IPv6 |

*Private Dualstack*

| Prefix    | Next Hop | Remark                           |
| --------- | -------- | -------------------------------- |
| 0.0.0.0/0 | NAT GW   | NAT for IPv4                     |
| ::/0      | EIGW     | Egress only for IPv6 + responses |

*Public IPv6-only*

| Prefix | Next Hop | Remark      |
| ------ | -------- | ----------- |
| ::/0   | IGW      | Public IPv6 |

*Private IPv6-only* (needs DNS64 feature on subnet level)

| Prefix       | Next Hop | Remark                           |
| ------------ | -------- | -------------------------------- |
| 64:ff9b::/96 | NAT GW   | NAT64 prefix for IPv4 targets    |
| ::/0         | EIGW     | Egress only for IPv6 + responses |

## Other Services

While IPv6 adoption is still rather slow, there have been significant advances over the last 12 months at AWS:

- [ALB/NLB support IPv6 now](https://aws.amazon.com/about-aws/whats-new/2021/11/application-load-balancer-network-load-balancer-end-to-end-ipv6-support/)
- [EKS can use IPv6](https://aws.amazon.com/about-aws/whats-new/2022/01/amazon-eks-ipv6/)
- [Lambda supports invocation from IPv6](https://aws.amazon.com/about-aws/whats-new/2021/12/aws-lambda-ipv6-endpoints-inbound-connections/)
- [The EC2 API is dual-stack now](https://aws.amazon.com/about-aws/whats-new/2021/01/amazon-ec2-api-supports-internet-protocol-version-6/)
- [Even Lightsail has IPv6 support](https://aws.amazon.com/about-aws/whats-new/2021/01/amazon-lightsail-supports-ipv6/)

## References

[^1]: [Stateful NAT64 (RFC6146)](https://datatracker.ietf.org/doc/html/rfc6146) 
[^2]: [IPv6 Addressing of IPv4/IPv6 Translators (RFC6052)](https://datatracker.ietf.org/doc/html/rfc6052)
[^3]: [DNS64: DNS Extensions for Network Address Translation (RFC6147)](https://datatracker.ietf.org/doc/html/rfc6147) 
