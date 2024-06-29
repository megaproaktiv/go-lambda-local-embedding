---
author: "Thomas Heinen"
title: "SBOMs on AWS - what?"
date: 2023-08-09
image: "img/2023/08/sbom-dreamstudio.jpg" # JPG because PNG won't compress below 1MB due to detail
thumbnail: "img/2023/08/sbom-dreamstudio.jpg"
toc: false
draft: false
categories: ["sbom", "aws"]
tags: ["sbom", "chef"]
---
Like most IT professionals, you might have read the title and googled "SBOM". Now that you know it stands for "Software Bill of Materials", read on to see why this will be very important in the next years. And what AWS can do to help you with this concept.

<!--more-->

## SBOMs and Log4J

An SBOM is simply a list of ingredients for your software solutions. Regardless of the code you have written, Lambda functions, Docker Containers, or HELM charts - every one of your solutions has external dependencies. If you think back to the infamous [Log4J bug in 2021](https://en.wikipedia.org/wiki/Log4Shell) you will likely remember everyone scrambling to find out where the vulnerable versions had been deployed.

Now imagine you already had SBOMs for all your deployments. It would have been a simple search for `log4j < 2.17.1` and you would have known. But as the saying goes, "Hindsight has 20:20 vision".

## Why You Need To Care About SBOMs

Due to the immense risks of undiscovered vulnerabilities in large IT ecosystems, there have been advancements over the last years under the term "Software Supply Chain Security" and various cyber resilience initiatives. One thing they all seem to stress is SBOMs.

As a result, the EU published the ["EU Cyber Resiliency Act"](https://digital-strategy.ec.europa.eu/en/library/cyber-resilience-act) in 2022, and in the US, the ["Executive Order 14028"](https://www.nist.gov/itl/executive-order-14028-improving-nations-cybersecurity/software-security-supply-chains-software-1) introduced the concept even before Log4Shell hit the world.

Both legislations will **require** software developers to create and provide an SBOM for their solutions. Even while the US initiative is only geared towards ISVs who provide software to the government, it is almost certain that this will extend into at least critical infrastructure like energy, transportation, etc.

While the respective texts are not mentioning it specifically, I would assume that even helper-Lambdas or custom OCI images deployed in consulting engagement might need SBOMs in the future - and be it only for customers of critical verticals.

## SBOM Formats and Types

The two main formats used for SBOMs are [SPDX](https://spdx.dev) and [CycloneDX](https://cyclonedx.org). While the US will also accept another standard, [SWID](https://csrc.nist.gov/projects/Software-Identification-SWID), the EU has decided to only allow these two formats for now.

In my opinion, CycloneDX is the more future-proof of these two: Not only does it support cryptographically signing SBOMs against tampering, but it also allows to specify any _used external APIs or even infrastructural dependencies_ like specific S3 buckets, Lambda functions, and much more.

AWS provides an [example for CycloneDX in the AWS Inspector documentation](https://docs.aws.amazon.com/inspector/latest/user/sbom-export.html):

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "version": 1,
  "metadata": {
    "timestamp": "2023-06-02T01:17:46Z",
    "component": null,
    "properties": [{
        "name": "imageId",
        "value": "sha256:c8ee97f7052776ef223080741f61fcdf6a3a9107810ea9649f904aa4269fdac6"
      },
      {
        "name": "architecture",
        "value": "arm64"
      },
      {
        "name": "accountId",
        "value": "111122223333"
      },
      {
        "name": "resourceType",
        "value": "AWS_ECR_CONTAINER_IMAGE"
      }
    ]
  },
  "components": [
    {
      "type": "library",
      "name": "pip",
      "purl": "pkg:pypi/pip@22.0.4?path=usr/local/lib/python3.8/site-packages/pip-22.0.4.dist-info/METADATA",
      "bom-ref": "98dc550d1e9a0b24161daaa0d535c699"
    },
    {
      "type": "application",
      "name": "libss2",
      "purl": "pkg:dpkg/libss2@1.44.5-1+deb10u3?arch=ARM64&epoch=0&upstream=libss2-1.44.5-1+deb10u3.src.dpkg",
      "bom-ref": "2f4d199d4ef9e2ae639b4f8d04a813a2"
    },
    // ...
  ]
}
```

Additionally to the technical format specifications, you can also classify [different types of SBOMs (PDF)](https://www.cisa.gov/sites/default/files/2023-04/sbom-types-document-508c.pdf):

- Source SBOMs: derived from the source code of your product
- Build SBOMs: created during the build process, including linked components
- Analyzed SBOMs: output of 3rd party tools which scan existing software
- Deployed SBOMs: a comprehensive list of all components of a system

## AWS Inspector

Starting June 2023, [AWS Inspector has received the capability to export SBOMs](https://aws.amazon.com/about-aws/whats-new/2023/06/software-bill-materials-export-capability-amazon-inspector/) into S3. It will scan EC2 instances, ECR repositories, and Lambda functions for their dependencies and export them as SPDX or CycloneDX into your S3 bucket.

As we are talking about AWS Inspector, we are not touching the Source SBOM or Build SBOM types in this case as the build process is already completed at this stage:

- EC2 scans will produce a "Deployed" SBOM of the whole system
- ECR/Lambda scans will create an "Analyzed" SBOM of the deployed packages

Both the German BSI and the US NITA discourage adding existing vulnerability information into SBOMs like AWS Inspector, but it is not a problem to do so. The argumentation is that while SBOMs are immutable for a specific software version, the information about vulnerabilities in the dependencies will change over time.

{{% notice info %}}
During my tests with AWS Inspector, I discovered that SBOM exports currently do not cover any Lambda Layers attached to your function. As a result, I had functions reporting no dependencies when I constantly used the pattern to deploy those as a base layer.<br>
<br>
Please keep an eye on your SBOMs, if this is a problem for you or if the issue has been fixed.
{{% /notice %}}

## Automating SBOM Exports

With a bit of code you can automate the export of SBOMs whenever new Lambda versions or ECR containers are deployed. You can connect to the `UpdateFunctionCode` event from Lambda to an EventBridge rule, transform the JSON, and then trigger Inspector on the updated resource via API Gateway.

![Architecture for automatic SBOM](/img/2023/08/sbom-architecture.png)

Sadly, there is no direct integration between EventBridge and Inspector yet, so you have to use API Gateway's direct service integration to creat this in a serverless (and Lambda-less) fashion.

## SBOM in CI/CD

Strictly speaking, AWS Inspector creates the SBOMs a bit late because it starts only after code has been deployed. If you want to tackle generation of the documents earlier, there is already a big collection of tools available to solve this.

Over the years, [syft](https://github.com/anchore/syft) has developed into one of the quasi-standards for SBOM generation, spanning a wide range of languages out of the box. Also, tools like [Snyk](https://snyk.io/blog/building-sbom-open-source-supply-chain-security/) can help you throughout the development lifecycle by creating and maintaining your software inventory. You can also use [TrendMicro's Artifact Scanner](https://cloudone.trendmicro.com/docs/container-security/tmas-about/) to not only scan OCI images, but also check vulnerabilities.

If you prefer tools which are specific to your language, there are a number of tools readily available - have a look at the [Awesome SBOM](https://github.com/awesomeSBOM/awesome-sbom) repository on GitHub.
