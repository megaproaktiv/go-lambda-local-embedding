---
title: "Scaling Down EKS Clusters at night"
author: "Benjamin Wagner"
date: 2023-08-04
toc: true
draft: false
image: "img/2023/08/scaling-down-eks-clusters-at-night.jpg"
thumbnail: "img/2023/08/scaling-down-eks-clusters-at-night-small.jpg"
categories: ["aws"]
tags:
  [
    "aws",
    "eks",
    "kubernetes",
    "costs",
    "keda",
  ]
---

Scaling down workloads at night or at the weekends is a common implementation task for companies building on AWS. By running only the applications that need to be available at any point in time, the total consumption of infrastructure resources can be reduced, and thus customers can benefit from the pay-by-use pricing models of cloud providers. 

<!--more-->

For EC2-based workloads, this is a fairly simple task: Using [Scheduled Scaling for Auto Scaling Groups](https://docs.aws.amazon.com/autoscaling/ec2/userguide/ec2-auto-scaling-scheduled-scaling.html), compute resources can be scaled up and down at a time schedule which is defined as a cron expression.

For EKS-based workloads, there is however an additional challenge. While usually one single application is running on an EC2 instance in the first setup, many application are running on the same EC2 instance when using a container orchestrator like kubernetes. Hence, simply shutting down an EC2 instance from the infrastructure side is not a viable solution for the following reasons:

* It will shut down random applications
* These applications will be rescheduled using managing controllers of the pods (e.g. a deployment)
* A cluster-autoscaler will spin up new EC2 instances to meet the demand

Instead, a solution is needed that will scale down the applications first, and the infrastructure will follow.

## Introducing Keda

Keda is an event-driven autoscaler for workloads on kubernetes. It has many integrations (so-called scalers) such as SNS, SQS or Kafka. For a full list, refer to [Keda Scalers](https://keda.sh/docs/2.11/scalers/). The operator implements a custom resource called *ScaledObject* which creates and updates Horizontal Pod Autoscalers (HPAs) inside the cluster. HPAs are updated by Keda whenever the event defined in the ScaledObject resource trigger.

Setting up keda is fairly simple: We only need to install the helm chart.

````
helm install keda kedacore/keda --namespace keda --create-namespace
````


## Solving the problem

To solve the problem described above, Keda's cron scaler can be utilized in order to scale applications (deployments) up and down based on time schedules.

The ScaledObject has a start and an end time, defined by cron expressions. At the start time, Keda will scale out the deployment to the desiredReplicas count. At the end time, it will scale back in to zero, unless a [minReplicaCount](https://keda.sh/docs/1.4/concepts/scaling-deployments/#scaledobject-spec) is defined. 

Refer to the [documentation](https://keda.sh/docs/2.0/scalers/cron/) to see how to use the cron scaler in a ScaledObject.


## Example Code

The following example will scale the deployment scalable-deployment-example to 5 replicas every morning at 7am and scale it back in to 1 replica at 7pm.

````
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: cron-scaledobject
  namespace: keda
spec:
  scaleTargetRef:
    name: scalable-deployment-example
  minReplicaCount: 1
  triggers:
  - type: cron
    metadata:
      timezone: Europe/Zurich
      start: "0 7 * * *"
      end: "0 19 * * *"
      desiredReplicas: "5"
````

---

Title Photo by [Benjamin Voros](https://unsplash.com/@vorosbenisop) on [Unsplash](https://unsplash.com/photos/U-Kty6HxcQc)