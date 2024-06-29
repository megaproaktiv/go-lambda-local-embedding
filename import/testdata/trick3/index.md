---
author: "Gernot Glawe"
title: "tRick: simple network 3 - Diversity (polyglott), Tooling, Fazit"
date: 2019-05-30
image: "img/2019/05/trick-overview.png"
thumbnail: "img/2019/05/trick-overview.png"
aliases:
    - /2019/trick-simple-3
toc: true
draft: false
tags: [devops,cloudformation,cicd,trick,iac]
categories: [AWS]
---

# Vergleich Infrastructure as Code (IaC) Frameworks - tRick Teil 3


### Alle Posts

1. [Abstraktion und Lines of Code](/2019/05/trick-simple-network-1-abstraktion-und-loc.html)
2. [Geschwindigkeit](/2019/05/trick-simple-network-2-geschwindigkeit.html)
3. [Diversity (polyglott), Tooling, Fazit](/2019/05/trick-simple-network-3-diversity-polyglott-tooling-fazit.html)

<!--more-->

## Diversity

|Framework      | Stars | Sum | Hcl | CloudFormation | Node | Python | Java | go | .NET |
|---            | ---   | --- | --- | ---            | ---  | ---    | ---  | ---| ---  |
| terraform     |  2    | 1   | ✅  | x              | x    | x      | x    | x  | x    |
| GoFormation   |  3    | 2   | x   | ✅             | x    | x      | x    | ✅ | x    |
| Pulumi        |  4    | 3   | x   | x              | ✅   | ✅     | x    | ✅ | x    |
| CDK           |  5    | 5   | x   | ✅             | ✅   | ✅     | ✅   | x  | ✅   |
| CloudFormation|  2    | 1   | x   | ✅             | x    | x      | x    | x  | x    |

(HashiCorp Configuration Language (HCL). )

## Tooling

Ideen

- Linting
- Static Code Checks
- Validation, Rulechecker
- Test of code (Unit Test)
- Integration Test
- Test of generated Language
- IDE plugins
- Graphical View



### terraform

{{< figure src="/img/2019/05/trick-overview-terraform.png" title="terraform" >}}

### goformation

{{< figure src="/img/2019/05/trick-overview-goformation.png" title="goformation" >}}

### pulumi

{{< figure src="/img/2019/05/trick-overview-pulumi.png" title="pulumi" >}}

### cdk

{{< figure src="/img/2019/05/trick-overview-cdk.png" title="cdk" >}}

### CloudFormation

{{< figure src="/img/2019/05/trick-overview-cloudformation.png" title="CloudFormation" >}}

## Fazit

Wenn ich die verschiedenen Tools betrachte, gibt es die große Unterscheidungen "Programmiernah" und "Konfigurationsnah".

Mein persönlicher Favorit bleibt das CDK, aber auch die anderen Tools haben mit der speziellen Mischung aus Features, Stabilität und Einfachheit ihre Fanbase.

Mein Tipp also: Ruhig einen Blick über den Tellerrand wagen und die anderen Tools probieren, um sich dann für einen Projekttyp für einen Standard zu entscheiden.