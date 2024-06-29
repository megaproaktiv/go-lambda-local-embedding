---
title: "Don't kill it with iron! How many pods can I start on an EKS node?"
author: "Benjamin Wagner"
date: 2022-10-31
toc: false
draft: false
image: "img/2022/10/eks-vpc-cni-banner.jpg"
thumbnail: "img/2022/10/eks-vpc-cni-banner.jpg"
tags: ["Container", "EKS", "Kubernetes", "Networking"]
summary: | 
    Every EC2 instance type has a limited number of ENIs and IP addresses that it can use. This can quickly cause EKS to not being able to schedule more pods on a node. Luckily, there is a simple solution for that.
---

<!-- Every EC2 instance type has a limited number of ENIs and IP addresses that it can use. This can quickly cause EKS to not being able to schedule more pods on a node. Luckily, there is a simple solution for that. -->

<span style="color:red">*0/5 nodes are available: 5 Too many pods*</span>

# The problem

Many of my clients have been facing this issue before: Pods cannot be scheduled in an EKS cluster whilst CPU and RAM utilization is often low. Lacking knowledge and time to research the topic, many decide to scale up the cluster by adding more nodes, incurring additional costs.

The origin of this issue is the number of IP addresses available on a kubernetes data plane node in AWS. Due to the default behaviour of the VPC CNI plugin (a.k.a. the aws-node daemonset), every pod gets assigned an IPv4 address from the subnets IP address space. From that, two problems can arise:

1. The subnet is running out of available IPv4 addresses. This can occur in poorly designed network setups or in very large enterprise environments. The solution for that is assigning secondary CIDR blocks for the VPC and configuring the EKS cluster to use [Custom Networking](https://docs.aws.amazon.com/eks/latest/userguide/cni-custom-network.html). It is the less likely problem to face and it's not covered in this article in more detail.
2. The EC2 instance can't claim more IPv4 addresses. Every EC2 instance type has a [limited number of ENIs and IP addresses](https://github.com/awslabs/amazon-eks-ami/blob/master/files/eni-max-pods.txt) that it can use. For example, a m6i.large instance with 2 vCPUs and 8GB RAM can use only 3 ENIs with 10 IPv4 addresses each. One address is claimed by the EC2 node itself, leaving 29 addresses to be used by pods running on that node. From my experience, this is a frequent problem and I'm going to summarize the solution in this blog post. Let's observe the problem in practice:

	```bash
	# Create an EKS Cluster
	eksctl create cluster --name my-cluster --without-nodegroup
	
	# Create a self-managed node group and set max-pods to a higher number
	# (here 1000)
	eksctl create nodegroup --cluster my-cluster --managed=false \
	--node-type m6i.large --nodes 1 --max-pods-per-node 1000 --name m6i-large
	
	# Test the default behaviour
	kubectl create deployment nginx --image=nginx --replicas=30
	kubectl get deployment -w
    
    ```
    
    As there are four pods running in the kube-system namespace and the total number of pods is 29, 25 pods (out of 30) will start and the remaining five will be pending. Describing one of them, we can see the following error:

	![EKS Pod FailedScheduling](/img/2022/10/eks_failed_scheduling.png)


# The solution

Luckily, there is a simple solution for that. Instead of attaching single IP addresses, it is [possible to assign /28 IPv4 prefix lists to the ENIs of EC2 instances](https://aws.amazon.com/about-aws/whats-new/2021/07/amazon-virtual-private-cloud-vpc-customers-can-assign-ip-prefixes-ec2-instances/). This can be used to (at least in theory) run around 16x as many pods on each EC2 node. For that, a few [configurations for the VPC CNI plugin](https://docs.aws.amazon.com/eks/latest/userguide/cni-increase-ip-addresses.html) are needed. For example:

```
# Delete the deployment to free addresses
kubectl delete deployment nginx  

# Configure VPC CNI to use /28 address prefixes
kubectl set env daemonset aws-node -n kube-system ENABLE_PREFIX_DELEGATION=true
kubectl set env ds aws-node -n kube-system WARM_PREFIX_TARGET=1
kubectl set env ds aws-node -n kube-system WARM_IP_TARGET=5
kubectl set env ds aws-node -n kube-system MINIMUM_IP_TARGET=2

# Wait for a new aws-node pod to come up and test again
kubectl create deployment nginx --image=nginx --replicas=100
kubectl get deployment -w
```

Cool, it works! We more than trippled the amount of pods on our node! Can we do more?

```
kubectl scale deployment nginx --replicas=500
kubectl get deployment -w
```

![EKS Node NotReady](/img/2022/10/eks_node_notready.png)

IPv4 addresses are not the limiting factor any more, but the node is breaking down. This seems reasonable given the amount of processes it is running. This leads us to another question though.

## How many pods *should* I run on a node?

AWS offers a simple [shell script for calculating the maximum number of pods](https://docs.aws.amazon.com/eks/latest/userguide/choosing-instance-type.html#determine-max-pods) that should run on a specific instance type. By default, it calculates the number of pods with the default behaviour of the VPC CNI:

```
./max-pods-calculator.sh --instance-type m6i.large --cni-version 1.9.0-eksbuild.1
29
```

Using a flag, we can calculate the number of pods with IP prefix lists enabled:

```
./max-pods-calculator.sh --instance-type m6i.large --cni-version 1.9.0-eksbuild.1  --cni-prefix-delegation-enabled
110
```

The `--max-pods-per-node` parameter in the first code snippet above should be set to this value. 1000 is not a good value. 

# The alternative solution

An alternative solution is [installing an alternate CNI plugin](https://docs.aws.amazon.com/eks/latest/userguide/alternate-cni-plugins.html) instead of the default VPC CNI plugin for cluster networking. This will result to IP addresses not being provided from the VPC CIDR, but from a virtual (potentially unlimited) address space. However, having another virtual IP address space within the environment will negatively impact transparency over your networking, and support for these alternative CNI plugins will not be provided by AWS.

While authoring this article, I experimented a bit with alternative CNI plugins and found that configuring them is fairly easy. However, the usage of IP prefix lists remediated the issue already, and thus there was no reason to further pursue this path. My strong recommendation goes for the VPC CNI plugin due to the abovementioned reasoning: transparency, ease and support.