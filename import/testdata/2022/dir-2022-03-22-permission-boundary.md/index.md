---
title: "Using Permission Boundaries to balance Security and Developer Productivity"
author: "Maurice Borgmeier"
date: 2022-03-22
toc: false
draft: false
image: "img/2022/03/permission_boundaries_venn.png"
thumbnail: "img/2022/03/permission_boundaries_venn.png"
categories: ["aws"]
tags:
  ["level-200", "iam", "permission-boundary", "security", "well-architected"]
summary: |
  There is a conflict between developer freedom and the requirements of security teams. In this post we'll look at one approach to address this tension: permission boundaries. They're an often overlooked part of IAM, but provide a valuable addition to our security toolkit.
---

There is an ongoing conflict between developer freedom and the security team's needs. Developers would like to have as much freedom to build secure solutions as possible, while the security team is responsible for ensuring solutions meet compliance needs. Many attempts try to solve this tension, and DevSecOps is a promising approach, in my opinion. Unfortunately, I have seen few real-world examples of that yet, so it's worth checking out one of the tools AWS gives us to address parts of this conflict. Today we'll talk about Permission Boundaries.

Before we do that, let me elaborate on the problem they help solve. In several projects we've consulted on, development and security teams don't work as closely together as they perhaps should. We have seen cases where the security team wants to review any IAM policy changes before they go into production. This is understandable and also a good practice. Unfortunately, this means that security teams become the bottleneck in release processes. Given the current market situation for qualified information security personnel, it's worth automating as much of this as possible.

One way to address this is permission boundaries. Let's quickly recap the core components of authentication and authorization in AWS to understand what they do. One way or another, a principal authenticates itself to AWS. Principals can be users, roles, or services. Authorization happens after the principals identify themselves. During authorization, all applicable policies are evaluated, and if there is no explicit deny for this API call **and** there is an Allow for it, the operation is permitted. These IAM policies govern which actions on specific resources will be allowed or denied.

One crucial step is often glanced over in this flow - it's the evaluation of all applicable policies. We usually deal with identity- and resource-based policies, both of which can grant or explicitly deny permissions. But these are not the only ones. Service Control Policies (SCPs) and permission boundaries filter the identity- and resource-based permissions. They only limit permissions and don't grant any on their own. This means even if an identity- or resource-based policy grants permission, it may still be denied if the service control policies or permission boundaries limit access.

![Venn Diagramm](/img/2022/03/permission_boundaries_venn.png)

What's the difference between service control policies and permission boundaries? SCPs apply to the whole AWS account and all resources within. SCPs are blunt instruments. On the other hand, permission boundaries are attached to individual identities and limit their privileges. They're more like a scalpel compared to SCPs. Here, we focus on permission boundaries. In practice, permission-boundaries are ordinary managed IAM policies attached to identities. This makes it possible to attach the same policy to multiple identities and thus centralize the management of said policy.

Why is this so useful? You can use permission boundaries in conditions. You can grant users or roles the permission to create other roles or users under the condition that they attach the permission boundary when the entity is created or updated. Policies can enforce the presence of the permission boundary. Breaking out of that sandbox will be impossible for your developers. They're free to create whichever role they want as long as they attach the permission boundary. You can rely on the fact that they won't ever be able to do more than that boundary allows.

Here is an example of an IAM policy that allows the creation of IAM users and roles under the condition that a specific permission boundary policy is attached:

```JSON
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "CreateOrChangeUserOnlyWithBoundary",
            "Effect": "Allow",
            "Action": [
                "iam:CreateUser",
                "iam:DeleteUserPolicy",
                "iam:AttachUserPolicy",
                "iam:DetachUserPolicy",
                "iam:PutUserPermissionsBoundary",
                "iam:PutUserPolicy"
            ],
            "Resource": "*",
            "Condition": {"StringEquals":
                {"iam:PermissionsBoundary": "arn:aws:iam::123456789012:policy/XCompanyBoundaries"}}
        },
		{
            "Sid": "CreateOrChangeRoleOnlyWithBoundary",
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:PutRolePolicy",
                "iam:AttachRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:DetachRolePolicy"
            ],
            "Resource": "*",
            "Condition": {
                "StringEquals": {
                    "iam:PermissionsBoundary": "arn:aws:iam::123456789012:policy/XCompanyBoundaries"
                }
            }
        },

	]
}
```

Permission boundaries are not a magic solution to all security problems. They define the maximum permission that could be granted, but that may not precisely mean the role has the least privileges it could have. Your developers should still limit the scope of the policies they write as much as possible. Permission boundaries act as another layer of defense against too broad permissions.

## Summary

In this post, we explored IAM permission boundaries and learned what they are, how they work, and what they can be used for. One of the use cases is to help the security team grant developers more freedom to build architectures within their AWS environment while limiting the blast radius their choices can have. Permission boundaries are not a silver bullet, but they are another valuable tool in our security toolbox.

Thank you to my colleague [André Reinecke](https://aws-blog.de/authors/andre-reinecke.html) for your interesting perspective on this topic!

Thank you for reading. For any questions, feedback or concerns, feel free to reach out to my via the channels mentioned in my bio.

&mdash; Maurice
