---
title: "Introduction to SSO with the IAM Identity Center and Entra ID"
author: "Maurice Borgmeier"
date: 2024-04-23
toc: false
draft: false
image: "img/2024/04/micah-williams-lmFJOx7hPc4-unsplash.jpg"
thumbnail: "img/2024/04/micah-williams-lmFJOx7hPc4-unsplash.jpg"
categories: ["aws"]
tags: ["level-200", "identity-center", "entra-id"]
summary: |
  IAM Identity Center, formerly AWS SSO, is often used as an access management solution in front of one or more AWS accounts. More often than not, its purpose is to grant access to AWS accounts within an AWS organization. Today we'll shed some light on the basic concepts and explain how this solution can be integrated with Azure AD which has recently been renamed to Entra ID to provide Single-Sign-On to your AWS environment.
---

IAM Identity Center, formerly AWS SSO, is often used as an access management solution in front of one or more AWS accounts. More often than not, its purpose is to grant access to AWS accounts within an AWS organization. Today we'll shed some light on the basic concepts and explain how this solution can be integrated with Azure AD which has recently been renamed to Entra ID to provide Single-Sign-On to your AWS environment.

As a user, your goal is to log into your AWS environment. This can be managed via Identity and Access Management for a single AWS account, but the real world is often more complicated. Managing access to sprawling AWS organizations is a challenge on its own, and this is where the IAM Identity Center helps you. It can integrate with a 3rd party identity provider such as your Entra ID and provide granular access to users and groups stored in that identity provider. You can think of Identity Center as a way to map AD users and groups to IAM Roles, just with extra steps. Let's dive in.

![Overview](/img/2024/04/idc_intro_overview.png)

To log into your AWS account, you first need to be authenticated, and in the subsequent authorization step, the Identity Center determines which roles you're allowed to assume in what AWS accounts. For this to work, the Identity Center needs to be set up in an account. Typically, it's associated with an AWS organization, but more on that later. In order for it to authenticate and authorize users, it needs to know about users and this is where the Identity Provider (IDP) enters the scene.

Identity Center has a built-in IDP, but you probably want to use your existing Entra ID to log in to the environment. That means we need to set up a connection, a mutual trust, between the Identity Center and Entra ID. This is commonly done through an Enterprise Application in Entra ID. The enterprise application gets assigned a subset of users and groups allowed to access the application. It can also be used to control which user attributes are available to the app.

![Authentication and User Sync](/img/2024/04/idc_intro_authentication_and_sync.png)

Next, we built up a trust relationship between the Identity Center and the Enterprise App by configuring the SAML protocol. The Identity Center can use this to authenticate users in Entra ID. If we leave it at that, authenticated users will be added to the internal Identity Store of Entra ID once they log in. This puts us in the awkward spot of having to wait for users to do something. Instead of doing that, we also configure SCIM 2.0 ([System for Cross-domain Identity Management](https://scim.cloud/)), a protocol that allows the enterprise app to proactively push all users and groups assigned to it to the Identity Center Identity Store. This protocol will be used to add, update, or remove users if they change on the Entra ID side of things. You get the option of choosing to sync only selected or all users and groups to the Identity Center.

The identity store maintains a reference to the original users and groups in Entra ID, an external ID combined with an external issuer. This allows it to track which Entra ID tenant a user or group originates from. Now, we're at the point where the Identity Center can authenticate users and already knows which users and groups have permission to access the Identity Center. This means we can now deal with the other problem: _authorization_.

To understand how the users and groups are now mapped to IAM roles, we have to talk about Permission Sets. You can think of a permission set like a template for an IAM role; it can have managed policies assigned to it, permissions boundaries, references to named policies, or inline policies. By itself, a permission set is harmless; it only starts doing something once an Assignment is added.

![Authorization](/img/2024/04/idc_intro_authorization_2.png)

The assignment couples a permission set with either a user or a group and an AWS Account. Once that's done, the Identity Center will provision an IAM role with the policies outlined in the permission set in the assigned account. You can spot these roles by the `AWSReservedSSO_` prefix in their name. That's basically all you need to do. Once you log in with your Entra ID account, you'll get a menu like this to pick the AWS Account + role to connect to.

![SSO Login](/img/2024/04/idc_intro_portal.png)

I should note that there's limited support for managing these resources through the SDK and CloudFormation at the time of writing this; aside from actually setting up and describing the integrations, you can do almost anything with the current SDK. For other things, you may have to rely on [undocumented API's](https://www.tecracer.com/blog/2024/04/using-undocumented-aws-apis-with-python.html) for now.

Regarding Infrastructure as Code (IaC), the resource coverage in CloudFormation is very limited. You can only manage Permission Sets and Assignments, and the latter becomes kind of a hassle in larger environments as you have to manage a lot of resources, and all user/group identifiers are UUIDs that are not very human-compatible.

(Note that these UUIDs are **not** the same as the Object IDs in Entra ID, but those are used as external IDs on the user and group resources.)

The pragmatic approach to IaC is to create the Permission Sets that way because those are relatively static and use well-understood resources such as different IAM policies that lend themselves to be versioned. Managing the assignments should be done another way. No one is going to thank you for making them edit UUIDs in a CloudFormation template all day when the UI makes this a lot easier.

This has been my brief introduction to the IAM Identity Center and its SSO with Entra ID. Hopefully, these concepts provide a good framework to dive deeper into the respective documentation. If you'd like to learn more about best practices or would like some support implementing this in your own environment, feel free to [get in touch with us](https://www.tecracer.com/en/contact/).

&mdash; Maurice


---

Title Photo by [Micah Williams](https://unsplash.com/@mr_williams_photography) on Unsplash