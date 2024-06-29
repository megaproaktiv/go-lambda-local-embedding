---
title: "Assigning EKS Namespaces to Node Groups"
author: "Benjamin Wagner"
date: 2023-08-25
toc: true
draft: false
image: "img/2023/08/assigning-eks-namespaces-to-node-groups.jpg"
thumbnail: "img/2023/08/assigning-eks-namespaces-to-node-groups.jpg"
categories: ["aws"]
tags:
  [
    "aws",
    "eks",
    "kubernetes",
    "operations"
  ]
---

In AWS EKS clusters, there are a couple of use cases for which all pods of a namespace should be automatically scheduled to specific nodes in Kubernetes, including:

* Clear allocation of data plane infrastructure (and costs) to teams in large organizations,
* Running critical workloads on on-demand nodes and not on spot nodes, or
* Using specific hardware, such as GPU, only by workloads that actually require it.

In this post, we will explore how to facilitate that in EKS.

<!--more-->

## The basic (and probably best) solution

There are several approaches to easily solve the problem at the deployment or pod level:

* NodeSelector
* Taints and Tolerations
* Affinity and Anti-Affinity

However, all of the aforementioned ways require that each application's Kubernetes manifests include the appropriate configuration, which can become effortsome in some cases. Therefore, there may be a desire to have Kubernetes do the mapping based on the namespace level and automatically adjust the applications' manifests. We will not go into the details of the above techniques in this post, but instead look into some advanced approaches to solve the problem.

## More elaborate solutions

In principle, solving the problem is possible with the [PodTolerationRestriction Admission Controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podtolerationrestriction) in kubernetes. Unlike the [PodNodeSelector](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector), this is mutating and not just validating, meaning it can change configurations to add a toleration for a tainted node to the pod. 

Unfortunately, [both Admission Controllers are unfortunately not supported in EKS](https://docs.aws.amazon.com/eks/latest/userguide/platform-versions.html). At the time of publishing this post, there are open Github issues ([here](https://github.com/aws/containers-roadmap/issues/304) and [here](https://github.com/aws/containers-roadmap/issues/739)) on the subject. So we need to find another solution.

The [MutatingAdmissionWebhook Admission Controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) is supported in EKS, i.e. a custom implementation of the desired logic is principally possible. This is especially helpful if the desired logic might be getting more complex than we assume in this blog post. Implementing this does however require a bit of work for something than in fact, plain kubernetes supports out of the box.

Alternatively, there is also an [open source project which implements a webhook for what we need](https://github.com/liangrog/admission-webhook-server). It doesn't seem to be actively maintained anymore though, judging from the latest commit date from December 2021.

## Tool-based solutions

Another option is to implement the manifest changes as part of a CI/CD pipeline, e.g. based on Github Action, Gitlab Runners or AWS CodePipeline. Some deployment tools such as ArgoCD might also support such manifest changes ([ArgoCD Overrides](https://argo-cd.readthedocs.io/en/stable/user-guide/parameters/)).

## Conclusion

As we have seen, there is no simple, out-of-the-box solution for the problem, which is particularly frustrating due to the fact the plain kubernetes does have such solution. It largely depends on the specific use case as well as the ecosystem tooling what approach is best. Whatever option is chosen, one should not forget about ongoing maintenance efforts for everything that is custom-built, and thus, the "manual" way of changing individual applications' manifest files is probably the best in many cases.


---

Title Photo by [Stéphan Valentin](https://unsplash.com/@valentinsteph) on [Unsplash](https://unsplash.com/photos/2w42JGUOuLs)