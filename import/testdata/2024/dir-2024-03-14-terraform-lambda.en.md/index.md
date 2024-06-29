---
author: "Maurice Borgmeier, Alexey Vidanov"
title: "Building Lambda with terraform"
date: 2024-03-14
image: "img/2019/05/terraform.png"
thumbnail: "img/2019/05/terraform.png"
toc: true
draft: false
tags: [devops,terraform,lambda,iac]
categories: [AWS]
---

{{% notice note %}}
Note: This is an updated version of [this blog](https://www.tecracer.com/blog/2019/05/building-lambda-with-terraform.html).
{{% /notice %}}

# Building Lambda Functions with Terraform

## Introduction

Many of us use Terraform to manage our infrastructure as code.

As AWS users, Lambda functions tend to be an important part of our infrastructure and its automation. Deploying - and 
especially building - Lambda functions with Terraform unfortunately isn't as straightforward as I'd like. (To be fair:
it's very much debatable whether you should use Terraform for this purpose, but I'd like to do that - and if I didn't,
you wouldn't get to read this article, so let's continue)

I'm going to divide the steps we need to take in order to deploy lambda functions into three parts: **Build**, **Compress** and **Use**.
Furthermore I'm going to call this our build pipeline, even though Terraform isn't really a build tool - you can't stop me.

Afterwards I'm going to show you two examples of build pipelines:
 1. A simplified version of the pipeline that only consists of the compress and use-steps
 1. ... and a more complex version with the actual Build-Step.

### Build

This step can involve a variety of different things, that depend on your use case and runtime environment of choice:
- Installation of dependencies (e.g. python packets)
- Running tests
- Compilation of code (e.g. go binaries)
- Setting up configuration files
- ...and more

For our purposes we're going to assume that it means running a script and continuing if that scripts' return code is 0.

### Compress

AWS Lambda only accepts zip files that contain the code alongside the Lambda handler for deployments - that's why we
need to provide a zip archive in our build pipeline.

### Use

Now that we've got our deployment package/artifact we're going to use it to create our lambda-function.

## Example 1 - Simplified Build Pipeline

In our simplified build pipeline we're going to skip the build step, which is not necessary, if you stick to the 
available libraries in the standard runtime environment.

We're going to start out with a directory structure that looks like this:

<!--
    Command for the tree-view
    tree -I 'venv|environments|switch_environment.sh|*.md|*.zip'
-->

```text
├── code
│   └── my_lambda_function
│       └── handler.py
├── lambda.tf
├── main.tf
├── permissions.tf
└── variables.tf

```

<!-- We're going to skip the IAM Role, because that's not very interesting --> 

This resource represents our *Compress* step - we're using the `archive_file` [data source](https://www.terraform.io/docs/providers/archive/d/archive_file.html)
of the archive provider (If you're using this for the first time in your project, you need to run `terraform init`
afterwards to initialize the provider).

Where you store the compressed zip-file (`output_path`) doesn't really matter - in any case I'd recommend you add the 
zip to your `.gitignore` file since there is no need to check in both the code and the build artifact. 

```hcl-terraform
data "archive_file" "my_lambda_function" {
  source_dir  = "${path.module}/code/my_lambda_function/"
  output_path = "${path.module}/code/my_lambda_function.zip"
  type        = "zip"
}
```

Now we can move on to the define step. In our function definition we're referencing a role that's not shown here -
replace it with your own. The `filename` argument points to the output of the above mentioned data source - our
compressed build artifact. The `source_code_hash` parameter references the sha256 hash of the zip archive and basically
makes sure that the code of the lambda function only gets updated, when the code in the repository (more specifically
the build output) is updated. 

```hcl-terraform
resource "aws_lambda_function" "my_lambda_function" {
  function_name    = "my_lambda_function"
  handler          = "handler.lambda_handler"
  role             = "${aws_iam_role.my_lambda_function_role.arn}"
  runtime          = "python3.11"
  timeout          = 60
  filename         = "${data.archive_file.my_lambda_function.output_path}"
  source_code_hash = "${data.archive_file.my_lambda_function.output_base64sha256}"
}
```

That's it - you should be able to modify the code and see the changes reflected in AWS after running `terraform apply`
(sometimes it takes a few seconds until you can see the new code in the console).

## Example 2 - Complete Version of the Build Pipeline

Our directory structure looks like this - you might be able to see the similarities:
```
├── code
│   └── my_lambda_function_with_dependencies
│       ├── build.sh
│       ├── handler.py
│       ├── package
│       └── requirements.txt
├── lambda.tf
├── main.tf
├── permissions.tf
└── variables.tf
```

The `build.sh` is very simple yet effective, it navigates to the scripts directory and installs all requirements into
the package directory.
```
#!/usr/bin/env bash

# Change to the script directory
cd "$(dirname "$0")"
pip install -r requirements.txt -t package/
```

The `handler.py` looks like this - it's a script that uses the `requests` library to find out the public IP of the
lambda function (or some proxy):

```python
# Tell python to include the package directory
import sys
sys.path.insert(0, 'package/')

import requests

def lambda_handler(event, context):

    my_ip = requests.get("https://api.ipify.org?format=json").json()

    return {"Public Ip": my_ip["ip"]}

```

Let's move to our build pipeline:

Our Build-Step is a null resource, that executes the `build.sh` whenever one of the following files change (see the
triggers-section):
- `handler.py`
- `requirements.txt`
- `build.sh`

```hcl-terraform
resource "terraform_data" "my_lambda_buildstep" {
  triggers_replace = {
    handler      = "${base64sha256(file("code/my_lambda_function_with_dependencies/handler.py"))}"
    requirements = "${base64sha256(file("code/my_lambda_function_with_dependencies/requirements.txt"))}"
    build        = "${base64sha256(file("code/my_lambda_function_with_dependencies/build.sh"))}"
  }

  provisioner "local-exec" {
    command = "${path.module}/code/my_lambda_function_with_dependencies/build.sh"
  }
}
```

The Compress-step is basically the same as above, with the exception of the `depends_on`-clause. Here we're waiting for
the build step to finish, before we compress the results.

```hcl-terraform
data "archive_file" "my_lambda_function_with_dependencies" {
  source_dir  = "${path.module}/code/my_lambda_function_with_dependencies/"
  output_path = "${path.module}/code/my_lambda_function_with_dependencies.zip"
  type        = "zip"

  depends_on = ["terraform_data.my_lambda_buildstep"]
}
```

Last but not least we use the resulting build artifact to deploy the lambda-function as described above.

```hcl-terraform
resource "aws_lambda_function" "my_lambda_function_with_dependencies" {
  function_name    = "my_lambda_function_with_dependencies"
  handler          = "handler.lambda_handler"
  role             = "${aws_iam_role.my_lambda_function_role.arn}"
  runtime          = "python3.11"
  timeout          = 60
  filename         = "${data.archive_file.my_lambda_function_with_dependencies.output_path}"
  source_code_hash = "${data.archive_file.my_lambda_function_with_dependencies.output_base64sha256}"
}
```

That's it, you should be able to run `terraform apply` and see the result in the console after a few seconds.

## References

This solution is inspired by a [discussion on Github](https://github.com/hashicorp/terraform/issues/8344), 
thanks to [@dkniffin](https://github.com/dkniffin) and [@pecigonzalo](https://github.com/pecigonzalo)