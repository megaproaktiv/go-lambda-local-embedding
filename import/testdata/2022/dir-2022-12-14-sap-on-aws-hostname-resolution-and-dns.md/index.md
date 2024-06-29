---
title: "Hostname Resolution and DNS with SAP on AWS"
author: "Fabian Brakowski"
date: 2022-12-14
toc: true
draft: false
image: "img/2019/09/fleur-dQf7RZhMOJU-unsplash.jpg"
thumbnail: "img/2019/09/fleur-dQf7RZhMOJU-unsplash.jpg"
categories: ["aws", "sap"]
tags: ["sap", "hostname resolution", "dns", "/etc/hosts", "Route53", "level-200", "terraform", "dhcp", "dhcp option sets", "/etc/resolvd"]
keywords:
    - sap
    - hostname resolution
    - dns
    - terraform
    - sap on aws automation
    
    
---

SAP systems running in a distributed environment have certain requirements regarding how to set the hostname and how those need to be resolved from other hosts. In our test landscape we use virtual hostnames to decouple the SAP instances from the underlying hardware which is running on a Red Hat Linux Server. This blog post will walk you through the components in AWS that fullfil those requirements and allow SAP instances to communicate while keeping administrative effort super low.

## Requirements

Within distributed SAP systems a lot of communication between servers (or better the installed SAP instances) is taking place. Traditionally, those instances rely heavily on the simple hostname (e.g. ``app-server``) instead of the FQDN (e.g. ``app-server.domain.name``). Therefore remote instances need to be reachable inside the local network using only their simple hostname.

Nowadays it's best practice to use virtual hostnames for each SAP instance. While this reduces some of the complexities related to the host's hostname, it introduces the need for secondary IP addresses with a properly resolved virtual hostname for each instance running on a host.

## /etc/hosts

In a typical On-Prem SAP environment, the (virtual) hostname resolution is often configured using the ``/etc/hosts`` file.
It allows for both forward and reverse DNS queries, however each entry has to be set individually in the file. Reverse DNS Lookup identifies which hostname belongs to a given IP address and is required by SAP.

A typical ``/etc/hosts`` file looks like the following:

```console
10.2.3.4 app-server.sap.landscape.domain.name app-server 
10.2.4.5 ascs.sap.landscape.domain.name ascs
```

This is easy and functional. However, it requires a lot of manual effort as it needs to be maintained on every host connected to a landscape.

## Setting up (Virtual) Hostname Resolution in AWS

What if we could replicate this same behaviour without any configuration on the host itself? Ideally this configuration is automatically maintained whenever a new server is launched, resulting in basically zero-maintenance.
In the following, I will guide you through the AWS components required for this setup.

### Secondary IP Addresses via Elastic Network Interface

Each virtual host requires its own dedicated IP address. **Elastic Network Interfaces** or ENIs enable EC2-Instances to communicate within your network. In this function they are also used to assign IP addresses to a server. By default, only traffic destined for a properly assigned IP address would be delivered to an instance. Other traffic, even if properly configured on the OS, is dropped.

For this reason a dedicated IP address must be assigned to the ENI for each SAP instance running on a server.

![Assign Secondary IP](/img/2022/09/sap-on-aws-secondary-ip-assign.PNG)

![Attached Secondary IP](/img/2022/09/sap-on-aws-secondary-ip.PNG)

The primary IP as well as the actual (non-virtual) hostname remain unused by the SAP server and can be used for administrative purposes.

### VPC Configuration: DHCP Option Set

As a one time activity, the **DHCP Option Set** needs to be properly maintained for each VPC. A VPC is a logically isolated network dedicated to your servers only. By default, every server launched into the VPC queries the VPCs DHCP server for information about its network configuration (e.g. the IP addresses previously discussed). A very important piece of information that the DHCP server provides, is the DNS configuration of the host. AWS allows us to set this information per VPC via DHCP Option Sets.

![Screenshot DHCP Option Set](/img/2022/09/sap-on-aws-dhcp-option-set.PNG)

To allow proper hostname resolution across servers, a common **domain name** needs to be set. Normally, customers want to use a custom name suiting their existing naming convention. Those domain names are configured via the DHCP Option Set. Whenever a server tries to resolve a hostname without specifying a domain, the server automatically attaches the domain name found in /etc/resolvd. This name is populated via DHCP Option Sets.

A second very important parameter that is set in DHCP Option Sets is the **domain name servers**. While the actual value depends on your overall network design (whether you use Route53 as the main DNS Service or whether you use an on premise DNS), for the proposed solution to work it is important that DNS queries eventually end up in AWS's DNS Service Route53. If you are unsure, keep the default value as this directly goes to Route53.

### Forward DNS Resolution

SAP Systems perform forward DNS resolution based on the simple hostname. The previous domain name setting ensures that the system automatically appends the domain name to the hostname if it can't resolve it. It then repeats the search.
After checking its cache, the OS will always first check in /etc/hosts. Afterwards it forwards the DNS query to the DNS server specified in /etc/resolvd.

The last required step for forward DNS resolution is a so called **private hosted zone in Amazon Route53**. A hosted zone specifies DNS entries for a single domain. During creation it can be specified as being private which requires the attachment of at least one VPC. DNS requests coming from that VPC are then able to query this hosted zone completely independent from the public DNS system.

![Screenshot Route53 Private Zone](/img/2022/09/sap-on-aws-route53_zone.PNG)

In that hosted zone, a separate entry is created for each virtual host in the system, mapping it to its own IP address. Afterward, forward DNS resolution works.

![Screenshot Route53 Entry](/img/2022/09/sap-on-aws-route53_entry.PNG)

![Screenshot Test via NSLOOKUP](/img/2022/09/sap-on-aws-nslookup.PNG)

### Reverse DNS Resolution

Reverse DNS queries work differently. Its goal is to provide a hostname based on an IP address. Reverse DNS uses so called PTR records that need to be inside a *.in-addr.arpa hosted zone. More details can be found [here](https://en.wikipedia.org/wiki/Reverse_DNS_lookup).
As a solution, we create this domain also as a private hosted zone and include our entries in the required format.

![Screenshot Route53 private zone for reverse DNS](/img/2022/09/sap-on-aws-route53-zone-reverse.PNG)

![Screenshot Route53 entry for reverse DNS](/img/2022/09/sap-on-aws-route53-entry-reverse.PNG)

![Screenshot Test Route53 entry for reverse DNS via nslookup](/img/2022/09/sap-on-aws-nslookup-reverse.PNG)

<!--- screenshot> <!--->

## Bringing it all together with Automation

All those steps seem more complicated than simply specifying DNS entries via ``/etc/hosts``. And in itself they are. However, being in AWS allows us to easily automate every single step. My personal tool of choice is Terraform.

I wrote a simple module that I use to create new EC2-Instances for SAP. As a parameter to that module I specify the servers virtual hostname. During ``terraform apply`` the instance is created which includes the  creation and assignment of a  secondary IP address. More importantly this IP address alongside the virtual hostname then gets written into Route53 which finalizes the setup.

```hcl
module "app_server" {
  source              = "./application_server"
  virtual_hostname    = "app-server"
  instance_type       = "m5.xlarge"
  domain_name         = "test-environment.sandbox.sap-on-aws.de"
  subnet              = module.network.subnet
}
```

Basically, after some initial settings that are valid for all servers, I do not have any extra work as the Terraform module will always take care of setting and if needed maintaining the values.


## Some relevant SAP Notes

- 611361 - Hostnames of SAP ABAP Platform servers
- 962955 - Use of virtual or logical TCP/IP host names
- 129997 - Hostname and IP address lookup
