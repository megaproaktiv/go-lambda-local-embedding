---
title: "Using AWS Security Hub for EKS Security"
author: "Benjamin Wagner"
date: 2023-08-04
toc: true
draft: false
image: "img/2023/08/using-security-hub-for-eks-security.png"
thumbnail: "img/2023/08/using-security-hub-for-eks-security.png"
categories: ["aws"]
tags:
  [
    "aws",
    "eks",
    "kubernetes",
    "security",
    "kube-bench",
  ]
---

[kube-bench](https://github.com/aquasecurity/kube-bench/tree/main) is a tool for checking kubernetes clusters against requirements defined in the [CIS Benchmark](https://www.cisecurity.org/benchmark/kubernetes). The tool runs locally on a kubernetes node, performs its checks and prompts the outputs to the shell or to files. This is quite unhandy, because it means that a user needs to pick up the logs, store them somewhere and analyze them. A deployment of the tool via kubernetes can ease the process for example with the `kubectl logs` command, but it is still far from perfect. Luckily, there is an integration in AWS Security Hub.

<!--more-->

[AWS Security Hub](https://aws.amazon.com/de/security-hub/) is a cloud-based service that does basically the same thing, but with a different scope. AWS Security Hub checks AWS accounts against best practices and regulations, and provides management capabilities investigate, trace and resolve findings. While Security Hub is a great tool when it comes to AWS security, it has no visibility inside kubernetes clusters by default. Integrating kube-bench with AWS Security Hub allows us to use Security Hub as the single pane of glass for our security posture without missing out the kubernetes view. kube-bench will simply send its findings to AWS Security Hub, so there is no manual log retrieval required any more.

## Integrating kube-bench and AWS Security Hub

There is a [tutorial video for installing kube-bench for EKS](https://www.youtube.com/watch?v=MwsUg3168YI) which guides through the setup process using the AWS console. It is a good starting point to get an understanding of how the integration works, but to use it at scale we need code, so here are some terraform snippets that you can copy and use to set it up in your AWS account(s). The code snippets are strongly aligned to the [example in the kube-bench docs](https://github.com/aquasecurity/kube-bench/blob/main/job-eks-asff.yaml), but refactored into terraform and extended by the AWS resources that we need.

First, we need to activate AWS Security Hub and subscribe to the (free) kube-bench product:

````
resource "aws_securityhub_account" "this" {}

resource "aws_securityhub_product_subscription" "kube_bench" {
  depends_on  = [aws_securityhub_account.this]
  product_arn = "arn:aws:securityhub:${data.aws_region.current.name}::product/aqua-security/kube-bench"
}
````

kube-bench needs permissions to write findings into Security Hub, so let's grant them. We are using an EKS [IAM Role for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) to provide access to the IAM role to the kube-bench pod.

````
resource "kubernetes_service_account" "kube_bench" {
  metadata {
    name      = "kube-bench"
    namespace = "kube-system"

    annotations = {
      "eks.amazonaws.com/role-arn" = module.kube_bench_irsa.iam_role_arn
    }
  }
}

module "kube_bench_irsa" {
  source  = "registry.terraform.io/terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "kube-bench"

  oidc_providers = {
    ex = {
      provider_arn               = module.k8s.oidc_provider_arn
      namespace_service_accounts = ["kube-system:kube-bench"]
    }
  }
}

data "aws_iam_policy_document" "kube_bench_irsa" {
  statement {
    actions = [
      "securityhub:BatchImportFindings"
    ]

    resources = [
      aws_securityhub_product_subscription.kube_bench.product_arn
    ]
  }
}

resource "aws_iam_role_policy" "kube_bench" {
  name   = "s3-readonly"
  role   = module.kube_bench_irsa.iam_role_name
  policy = data.aws_iam_policy_document.kube_bench_irsa.json
}
````

Lastly, we need to create a ConfigMap that kube-bench will use to push the findings into Security Hub:

````
resource "kubernetes_config_map" "kube_bench" {
  metadata {
    name      = "kube-bench-eks-config"
    namespace = "kube-system"
  }

  data = {
    "config.yaml" = <<-EOF
    AWS_ACCOUNT : "${data.aws_caller_identity.current.id}"
    AWS_REGION : "${data.aws_region.current.name}"
    CLUSTER_ARN : "${module.k8s.cluster_arn}"
    EOF
  }
}
````

## Running kube-bench

We run kube-bench using a kubernetes Job:

````
resource "kubernetes_job" "kube_bench" {
  metadata {
    name      = "kube-bench"
    namespace = "kube-system"
  }
  spec {
    template {
      metadata {}
      spec {
        host_pid = true
        container {
          name  = "kube-bench"
          image = "docker.io/aquasec/kube-bench:latest"
          command = [
            "kube-bench",
            "run",
            "--asff",
            "--logtostderr",
            "--v",
            "3"
          ]
          env {
            name = "NODE_NAME"
            value_from {
              field_ref {
                field_path = "spec.nodeName"
              }
            }
          }
          volume_mount {
            name       = "var-lib-kubelet"
            mount_path = "/var/lib/kubelet"
            read_only  = true
          }
          volume_mount {
            name       = "etc-systemd"
            mount_path = "/etc/systemd"
            read_only  = true
          }
          volume_mount {
            name       = "etc-kubernetes"
            mount_path = "/etc/kubernetes"
            read_only  = true
          }
          volume_mount {
            name       = "kube-bench-eks-config"
            mount_path = "/opt/kube-bench/cfg/eks-1.2.0/config.yaml"
            sub_path   = "config.yaml"
            read_only  = true
          }
        }
        restart_policy       = "Never"
        service_account_name = kubernetes_service_account.kube_bench.metadata.0.name
        volume {
          name = "var-lib-kubelet"
          host_path {
            path = "/var/lib/kubelet"
          }
        }
        volume {
          name = "etc-systemd"
          host_path {
            path = "/etc/systemd"
          }
        }
        volume {
          name = "etc-kubernetes"
          host_path {
            path = "/etc/kubernetes"
          }
        }
        volume {
          name = "kube-bench-eks-config"
          config_map {
            name = kubernetes_config_map.kube_bench.metadata.0.name
            items {
              key  = "config.yaml"
              path = "config.yaml"
            }
          }
        }
      }
    }
  }
  wait_for_completion = false

  depends_on = [
    aws_securityhub_product_subscription.kube_bench
  ]
}
````

For regular checking of real-world (not playground) environments, a kubernetes CronJob can be used instead.

## Using Security Hub to see the results

In AWS Security Hub, under _Findings_, we can now see a list of findings. Let's set a filter on the kube-bench findings:

`Product name is Kube-bench`

![kube-bench finding list in AWS Security Hub](/img/2023/08/using-security-hub-for-eks-security.png)

When browsing through the findings, it is striking that many of them have _(Manual)_ in their names, and only a few have _(Automated)_ in their names. As the words indicate, only the the _(Automated)_ checks are performed by kube-bench and the others are always pushed into Security Hub. This is a bit unfortunate as the tool doesn't provide the amount of automation as one could assume.

Opening one of the findings, we can see more details, however it doesn't show us which kubernetes resource failed the check. I would have expected more visibility here as well. Still, the tool gives good guidance into how to trace down / resolve the findings. Using the _Workflow Status_ field, we can document which findings have been notified (e.g. to administrators), surpressed or resolved.

![kube-bench finding details in AWS Security Hub](/img/2023/08/using-security-hub-for-eks-security-edit-finding.png)

## Conclusion

While playing around with tool, it seemed to me that kube-bench is not that mature yet. For example, the documentation is rather poor and the degree of automation of the checks is improveable. On the positive side, the integration in AWS Security Hub works really seamlessly, and it is a great benefit to have a tool that will automatically include the latest CIS Benchmark checks. Providing visibility into the best practices defined in the CIS Benchmark is the starting point for improving kubernetes cluster security, and the management capabilities of AWS Security Hub facilitate a structured remediation process of these findings. I am looking forward of the future developments of both tools!

---