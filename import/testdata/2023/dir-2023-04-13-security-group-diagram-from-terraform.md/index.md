---
title: "Visualize AWS Security Groups and Rules from Terraform State"
author: "Fabian Brakowski"
date: 2023-04-13
toc: false
draft: false
image: "img/2023/04/security-group-diagram-from-terraform-result.png"
thumbnail: "img/2023/04/security-group-diagram-from-terraform-result.png"
categories: ["aws", "terraform"]
tags: ["terraform", "diagrams"]
---
In an ever-changing AWS environment, maintaining manually created architecture diagrams can be challenging. For this reason, I decided to find an automated solution for generating diagrams based on the Terraform State File. In this blog post, I will introduce a tool that simplifies the maintenance of architecture diagrams, ensuring their accuracy and providing a clear understanding of AWS security groups and their interactions.<!--more-->

### Issue

The challenge is to demonstrate how multiple security groups interact with each other. Manually maintaining diagrams with tools like draw.io can be time-consuming and may lead to inconsistencies and errors, which can impact overall security.


## Introducing the Python Solution for Visualizing AWS Security Groups and Rules

Inspired by [Gernot's blog post about "diagrams as code" on AWS with CDK and D2](/2023/04/a-new-simple-approach-to-diagram-as-code-on-aws-with-cdk-and-d2.html), I developed a Terraform-specific solution using Python to parse the Terraform State File and generate diagrams. The full code can be found on [GitHub](https://github.com/tecracer/aws-security-group-diagram-from-terraform).

### How It Works

1. Parse the Terraform State File to extract security group information as well as prefix lists and hardcoded IP ranges.
2. Store the data in a custom format for easier diagram creation.
3. Generate the diagram using the extracted data and a diagram generation library.

## Demonstration

The following example demonstrates the effectiveness of this solution, using sample Terraform code for a cross-account setup.

1. First, clone the GitHub repository containing the Python tool:

```bash
git clone https://github.com/tecracer/aws-security-group-diagram-from-terraform.git
```

2. Next, navigate to the repository folder and install the required dependencies:

```bash
cd aws-security-group-diagram-from-terraform
pip install -r requirements.txt
```

3. Now, from the directory of your terraform project run:

```bash
terraform show -json | python3 /path/to/aws-security-group-diagram-from-terraform/main.py -i --output_filename fancy_diagram --output_format png
```

This will directly read from the state file and will generate a diagram with the specified file name and format.

Open the generated diagram file to review the visual representation of your AWS security groups and rules.

![Created Diagram](/img/2023/04/security-group-diagram-from-terraform-result.png)

In the current version, the diagram may not look perfect, but there is much room for adjustments as the `diagrams` library offers many ways in which to customize the output file.
Note that the icons used are not the actual icons for security groups as they are not released yet for the library, but changing the icon is easy and can be done later.

## Conclusion

This Python tool offers a practical solution for visualizing AWS security groups and rules directly from the Terraform State File, saving time and effort while improving the accuracy and maintainability of architecture diagrams. Give it a try and let me know how it works for you!