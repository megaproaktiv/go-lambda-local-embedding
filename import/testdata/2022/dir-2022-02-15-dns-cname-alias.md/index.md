---
title: "How ALIAS records can reduce initial load times for your website"
author: "Maurice Borgmeier"
date: 2022-02-15
toc: false
draft: false
image: "img/2022/02/r53_hierarchy.png"
thumbnail: "img/2022/02/r53_hierarchy.png"
categories: ["aws"]
tags: ["level-200", "route53", "dns"]
summary: |
    DNS is a core component of the internet. In this post we'll briefly take a look at how it works and what the difference between CNAME and ALIAS records in Amazon Route53 are.
---

“It’s always DNS” is what you’ll commonly hear during speculations about the cause of a large-scale outage of major websites. It’s true in surprisingly many cases, so it’s worth understanding DNS in a bit more detail. Today we’ll take a look at two commonly used features of AWS Route53, what they do, how they work, and when to use which. We’ll talk about CNAME and ALIAS records.

First, the basics so we’re on the same page. The Domain Name System translates human-readable names into computer-understandable IPs. It’s an integral part of what makes the internet usable for us as humans because nobody wants to remember 12-digit numbers for IPv4, and I’m not even getting into IPv6 here. For a simple mental model, you can assume it’s a key-value store where the key is a domain name, and the value is a list of one or more IP addresses that you get when you query the name.

Reality is a bit more complicated, however. The ideas of avoiding single points of failure and being resilient to outages in parts of it are core to the internet. So, a single central entity managing these mappings is not a good idea, mainly because it wouldn’t scale. That’s why there is a hierarchy to DNS. A typical domain name has different parts: “connect.tecRacer.de.” consists of multiple components - the dot, in the end, is intentional, by the way.

The dot, in the end, refers to the[ root zone](https://en.wikipedia.org/wiki/DNS_root_zone) - a set of 13 [geographically dispersed](https://root-servers.org/) DNS Servers that manage the names of the next level. In reality, there are more than 13 of these, but that’s an implementation detail. The servers in the root zone manage the top-level domains ([TLDs](https://en.wikipedia.org/wiki/Top-level_domain)); in our case, that’s the “.de” part. There’s another set of servers responsible for anything in front of the “.de.” The servers in the root zone delegate ownership and the responsibility to serve requests to them using NS-Records (Name-Server-Records). So NS records are primarily used to configure which name servers manage a [domain](https://en.wikipedia.org/wiki/Domain_name).

![Root Zone](/img/2022/02/r53_root_zone.png)

The name servers for the .de TLD know about all domains belonging to the TLD, such as tecRacer.de. They have an NS record for each domain that points to the DNS servers responsible for that domain. The tecRacer.de DNS-Servers are now responsible for answering questions about connect.tecRacer.de or where users can fit the website. Here is where Route53 comes in.

Route53 is AWS’ DNS-Service, and in our case, the tecRacer.de domain translates to a so-called public hosted zone in Route53. Here the records for tecRacer.de and its subdomains are managed. Now we come back to our original question about CNAME and ALIAS records. CNAME or canonical name is a standardized type of record that you can put in your hosted zone.

![DNS Hierarchy](/img/2022/02/r53_hierarchy.png)

CNAME records act as a pointer to another domain. It can say something like, “Oh, you want to get to connect.tecRacer.de, then you have to look up the IP address of load-balancer-123.elb.eu-central-1.amazonaws.com”. Why would you do that? Sometimes, a single server is responsible for multiple subdomains, and in that case, you don’t want to update all subdomain records if the IP of the Webserver on the base domain changes. You add all those subdomain records as CNAMEs that point to the base domain to avoid that.

This behavior is transparent to the end-user. Your browser will handle this in the background for you. It’s something the DNS-resolver of your operating system does when the browser asks it to return an IP address for a given domain. A-Records (IPv4) or AAAA-Records (IPv6) store IP addresses. The resolver will request the DNS records for the domain and sees that it received a CNAME record instead of an IP address (A-Record), so it does another lookup for the record in the CNAME and returns its IP address. Lookups can go on in a loop for a while if multiple levels of CNAMEs point to each other.

ALIAS records do something similar but in a different way. They are not exposed to clients of your DNS server because they are not part of the official standard. Effectively they allow you to set up a pointer from your domain to some supported AWS resources without a CNAME record. An ALIAS record will look like a typical A-Record to your client because it immediately returns the set of IP addresses for your load balancer, CloudFront domain, and others. It potentially saves you one network round-trip and speeds up DNS resolution.

![CNAME vs ALIAS](/img/2022/02/r53_cname_alias.png)

So is using ALIAS over CNAME records whenever possible is a no-brainer? For the most part, yes, but you shouldn’t overestimate its significance. DNS is (in)famous for caching. That’s part of what makes it fast, but in many cases also the reason why “it’s always DNS.” Caching of lookup information is performed on many levels, including your local resolver. The first lookup will probably be a little bit faster if you’re using ALIAS records because there is one round-trip less. Subsequent lookups will be cached on multiple levels so that it won’t make much of a difference.

ALIAS records are only usable when Route53 manages your domain. They’re beneficial, but it’s not necessarily worth migrating your domain to Route53 to be able to use them. CNAMEs work fine as well. The ability to manage everything through a single API may be a stronger motivation, which Route53 makes possible.

In this post, we’ve learned about the difference between CNAME and ALIAS records, as well as a little bit more detail about how DNS works. Hopefully, you gained something from it. For any questions, feedback, or concerns, feel free to reach out to me via the social media channels linked in my bio.

&mdash; Maurice
