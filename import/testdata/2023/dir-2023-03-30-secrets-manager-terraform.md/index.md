---
title: "Enhancing Security in Terraform with AWS Secrets Manager"
author: "Alexey Vidanov"
date: 2023-03-30
toc: true
draft: false
image: "img/2023/03/ente-av-tecracer-2023-03.jpg"
thumbnail: "img/2023/03/ente-av-tecracer-2023-03.jpg"
categories: ["aws"]
tags: ["terraform", "secrets manager", "security", "IaC", "AWS"]
---

Keeping track of your passwords is already challenging in your personal life. It can be more difficult when you want to build and deploy secure applications in the cloud. Today we’ll show you a few ways of managing secrets in your Terraform deployment. We’ll teach you about common pitfalls like the random_password resource and more appropriate alternatives.
<!--more-->

A production-ready secret store typically has strong encryption to protect secrets at rest, as well as robust access controls to prevent unauthorized access to those secrets. Additionally, it should have reliable backup and recovery mechanisms to ensure that secrets can be restored in the event of a failure or disaster.

Another important aspect of a stable secret store is the ability to audit all access and changes to the stored secrets. This means keeping a detailed log of who accessed which secrets and when, as well as tracking any modifications or additions to the store. Having a comprehensive audit trail allows for easier troubleshooting, compliance with regulatory requirements, and identification of potential security breaches or unauthorized access attempts. The audit logs should be securely stored and easily accessible to authorized personnel for review and analysis.

Securely managing secrets in Terraform can be challenging. This article explores an alternative approach to using `random_password` for generating and storing secrets in AWS Secrets Manager, while maintaining high security standards.

## AWS Secrets Manager: A Common Solution

A widely adopted approach is to rely on AWS Secrets Manager for storing sensitive data. AWS Secrets Manager offers several key benefits for handling sensitive information:

1. **Centralized secret storage and access control**: Secrets Manager provides a secure location for storing and managing secrets. Using AWS Identity and Access Management (IAM), you can define granular policies to control access, ensuring that only authorized users and applications have access to sensitive information.
2. **Automated secret rotation**: Secrets Manager supports automatic secret rotation, enabling you to periodically generate new secrets and automatically update dependent services without manual intervention (if the service is compatible), reducing the risk of secrets being compromised.
3. **Auditing and encryption**: Integration with AWS CloudTrail allows you to track and audit all actions related to your secrets. Secrets Manager encrypts your secrets using AWS Key Management Service (KMS) keys, ensuring that sensitive information is protected both at rest and in transit.

Utilizing these advantages, AWS Secrets Manager delivers a powerful and streamlined solution for managing sensitive information within your Terraform infrastructure.

## The Drawbacks of Using random_password

Securely handling secrets like passwords, API keys, and tokens when working with Terraform and AWS Secrets Manager can pose challenges. It may be tempting to use Terraform's `random_password` resource and the random provider to generate and store secrets in AWS Secrets Manager. This approach comes with significant security risks, primarily that the secret is saved in the Terraform state file as plaintext. This may expose sensitive information to unauthorized access and undermines the security advantages provided by AWS Secrets Manager, such as encryption at rest and in transit, fine-grained access control, and centralized secret storage. As a result, it is crucial to seek alternative methods that maintain the desired security standards while leveraging the benefits of AWS Secrets Manager.

## A Secure Alternative: Using `terraform_data` and local-exec Provisioner

![Terraform and AWS Secrets Manager](/img/2023/03/terraform-aws-secrets-manager.png)

You can generate and store a secret in AWS Secrets Manager using the [terraform_data](https://developer.hashicorp.com/terraform/language/resources/terraform-data#example-usage-null_resource-replacement) and local-exec provisioner, to avoid saving the sensitive information in the Terraform state file. Let's dive into the solution and discuss how it works.

{{% notice note %}}
The terraform_data resource was introduced in the Terraform Version 1.4. If you use an older version, you can use the null_resource instead.
{{% /notice %}}

In this example we store a JSON object with a user and a generated password as secret. The comments within the code describe the key functionality, providing an understanding of steps involved.

```hcl

################################################################################
# secrets.tf
################################################################################

provider "aws" {
  region = "eu-central-1"
}

variable "master_user_name" {
  description = "The variable for the user name"
  default     = "user"
}

resource "aws_secretsmanager_secret" "this" {
  name_prefix = "/AwesomeApp/UserPassword"
  tags = {
    owner          = "tecracer"
    environment    = "dev"
    technical-name = "tf-secrets-demo"
  }
}

resource "terraform_data" "create_secret_version" {
  triggers_replace = [
    aws_secretsmanager_secret.this.id
  ]
  provisioner "local-exec" {
    command     = <<-EOT
      # Generate a random 16-character base64-encoded password
      PASSWORD=$(openssl rand -base64 16)
      # Create a JSON object with username and password
      SECRET_JSON=$(echo '{"username": "${var.master_user_name}", "password": "'$PASSWORD'" }')
      # Store the JSON object in AWS Secrets Manager
      aws secretsmanager put-secret-value \
        --secret-id ${aws_secretsmanager_secret.this.id} \
        --secret-string "$SECRET_JSON"
    EOT
    interpreter = ["/bin/sh", "-c"]
  }
  depends_on = [aws_secretsmanager_secret.this]
}
```

Please note, that by default, only the current version and the latest version are stored. To have a more advanced secret store, the following things would be important:

- **Versioning:** Secret versions should be tagged with a named alias, such as a timestamp, to prevent accidentally destroying the current environment with the changes in the Terraform code.
- **Rotation:** Defining the secrets rotation and a recovery window.
- **Permissions policy:** Attaching a permissions policy to an AWS Secrets Manager secret.
- **KMS Key:** The AWS KMS key to be used to encrypt the secret values in the versions stored in this secret. If you don't specify this value, then Secrets Manager defaults to using the AWS account's default KMS key (the one named `aws/secretsmanager`). If the default KMS key with that name doesn't yet exist, then AWS Secrets Manager creates it for you automatically the first time.
-  **CloudTrail logs:** Auditing the monitoring with CloudTrail logs.

This solution effectively generates and stores a secret in AWS Secrets Manager without saving it in the Terraform state file. 

![aws-secrets-manager](/img/2023/03/aws-secrets-manager.png)

## Accessing the Generated Secret

Retrieving and utilizing secrets in Terraform presents a challenge, as if you use [data sources](https://developer.hashicorp.com/terraform/language/data-sources), you will store secrets in the Terraform state file, which is undesirable. Fortunately, if you have created secrets for use with AWS services, native support is available for some of them. These are Amazon RDS, AWS Lambda, EC2, ECS for example. This enables you to use AWS Secrets Manager secret by ID without having to retrieve the actual secret values.

If the AWS service doesn't offer native support for secret handling, or you need to use secrets with non-AWS services, the most effective method for accessing secrets without storing them in the state file is by using environment variables.

It's important to note that you cannot directly export an environment variable from Terraform code during execution. Terraform operates as an independent process, and any environment variables set within the Terraform process won't be passed back to the parent shell.

A viable workaround is to create a wrapper script that first sets the environment variable using a shell command and then invokes Terraform with the configured environment variable. The example provided demonstrates how to create a wrapper script using Bash for Linux or Mac OS X systems. If you are using Windows, you will need to create a similar script using PowerShell. The primary use case for this code is to integrate the Terraform configuration within a CI/CD pipeline, which typically runs on Linux-based runners.

1. Create a file called `run-terraform.sh` with the following content:

   ```bash
   #!/bin/bash
   
   # Search the secret by tags, get the secret value and store it in an environment variable
   MY_SECRET_ID=$(aws secretsmanager list-secrets --query "SecretList[?Tags[?Key=='technical-name' && Value=='tf-secrets-demo'] && Tags[?Key=='environment' && Value=='dev']].{ID:ARN}" --output text)
   export MY_SECRET=$(aws secretsmanager get-secret-value --secret-id "$MY_SECRET_ID" --query 'SecretString' --output text)
   # Call Terraform commands with the environment variable
   terraform init
   terraform apply -var "MY_SECRET=$MY_SECRET"
   ```

2. Make the script executable:

   `chmod +x run-terraform.sh`

3. Run the script:

   `./run-terraform.sh`

4. To utilize the secret within the previously created code, you can refer to the following code snippet:

```hcl
################################################################################
# access-secrets.tf
################################################################################
variable "MY_SECRET" {
  description = "The variable for the user name and passw"
}
locals {
  username = jsondecode(var.MY_SECRET).username
  password = jsondecode(var.MY_SECRET).password
}
# Then you can use the locals in your code by referring them like this:
# local.username
# local.password
# but to keep your state file free of the sensitive secrets
# you can use these locals again only in the provisioner "local-exec"
# for the resource creating you can use placeholder passwords
# and then replace them with the right secrets using terraform_data
```

## Conclusion

While this approach might seem unconventional, it offers an extra layer of security for sensitive information in Terraform projects.

When tasked with ensuring security and leveraging the powerful features of Terraform and AWS, it's vital to learn about and implement strategies that maintain the integrity of sensitive information. This example illustrates an effective approach for securely managing secrets within your infrastructure.

I hope you had fun and learned something new while working through this short example. I am looking forward to your feedback and questions. If you want to take a look at the complete example code please visit my [Github](https://github.com/vidanov/tecracer-blog-projects/tree/main/secrets-terraform).

— [Alexey](https://www.linkedin.com/comm/mynetwork/discovery-see-all?usecase=PEOPLE_FOLLOWS&followMember=vidanov)

P.S. If you're looking to expand your knowledge and skills in Terraform and AWS, consider enrolling in our training course: "[Terraform Engineering On AWS.](https://www.tecracer.com/training/terraform-engineering/)" This comprehensive course covers various aspects of using Terraform with AWS, enabling you to confidently design, deploy, and manage infrastructure in the cloud. For more information and to register for the course, please visit [our training website](https://www.tecracer.com/training/terraform-engineering/).

 
