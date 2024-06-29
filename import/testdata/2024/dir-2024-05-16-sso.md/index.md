---
title: "Switching Identity Providers in the IAM Identity Center"
author: "Maurice Borgmeier"
date: 2024-05-16
toc: false
draft: false
image: "img/2024/05/idc_migration_julia-craice-faCwTallTC0-unsplash.jpg"
thumbnail: "img/2024/05/idc_migration_julia-craice-faCwTallTC0-unsplash.jpg"
categories: ["aws"]
tags: ["level-200", "identity-center", "entra-id", "migration"]
summary: |
  Switching the Identity Provider in the IAM Identity Center while keeping all permissions intact and minimizing disruptions can be a daunting task.
  In this post I'm explaining how we solved this for one of our customers.
---

Mergers and acquisitions are usually decided at the top of the corporate hierarchy and eventually trickle down to regular technologists who need to implement changes. Today, I'm sharing one of these stories.

One of our customers was acquired by another company a while ago, and they are in the middle of integrating their identity management solutions. In a nutshell, that means all users and groups that our customer has in their Entra ID (formerly known as Azure AD) are recreated in the Entra ID Directory of the parent company. This is relevant to us because access to AWS accounts is granted based on a connection to the old Entra ID.

The customer uses the IAM Identity Center (formerly AWS SSO) to authenticate users in their multi-account structure. They map users and groups from their Entra ID to IAM roles in their organization via Permission Sets. If you're not familiar with the terminology, I recommend you check out [this blog post introducing the Identity Center](https://www.tecracer.com/blog/2024/04/introduction-to-sso-with-the-iam-identity-center-and-entra-id.html).

The challenge we faced was migrating from their existing Entra ID (old) to the parent companies' Entra ID (new) as an identity provider in the IAM Identity Center while minimizing disruptions to DevOps teams operating their services and retaining the current permission structure. Minimizing disruptions is vital because this solution manages access to all AWS accounts. If that's disrupted, service teams won't be able to access their environments to solve problems.

The actual services within these accounts should be fine as the Identity Center manages humans accessing AWS accounts and not the communication of the systems within. Nevertheless, we want to be cautious.

## Constraints

There are a few constraints that you need to know about when approaching something like this:

1. IAM Identity Center only supports a single active identity provider at any point in time
2. You can't update the external IDs and issuers on users and groups in the Identity Center identity store

These constraints are important because the first means you can't completely avoid disruptions. There will be a short downtime when you switch the identity providers in the IAM Identity Center. Arguably, the second is more problematic because it means you'll have to migrate your permission structure, too, since new users and groups will be created, and there's no direct way to map them to the old users and groups.

## Our Approach

Aside from the technical implementation, lots of communication is required in advance of and during the migration to ensure everyone is aware of and prepared to deal with the changes. While this is critical work, it's not the focus of this document. We'll focus on the technical aspects, which start with preparing for the changes.

### Preparation

In preparation for the cutover, we create files that map old users to new users and old groups to new groups based on their Entra ID object identifiers. This can be done without impacting the existing system, and having them available during cutover is essential. Additionally, we create an IAM user with permissions to manage the identity center because once we switch over, our Entra ID-based users won't work immediately, and we need to have access to the environment. Furthermore, we export the existing permission structure, i.e., the permission sets and their assignments to accounts, users, and groups. Lastly, we will create a new Enterprise app in the new Entra ID and assign new users and groups to it.

![Preparation](/img/2024/05/idc_migration_preparation.png)

### Cutover

With all these preparations done, we can start the cutover during a prearranged time window. During the cutover, we replace the existing IDP in the Identity Center and set up SAML authentication as well as SCIM for user provisioning with the new Enterprise app. Next we have to wait a few minutes for SCIM to provision all users and groups in the Identity Center identity store. Once that's done, we use the same script as we did during the preparation stage to export the permission structure with permission sets, users, groups, and all the assignments.

![Post Cutover](/img/2024/05/idc_migration_post_cutover.png)

The next step happens offline. We merge the old export with the new export using the aforementioned old user/group to new user/group mappings and update the identifiers in all assignments to point to the new users/groups instead of the old ones. We specifically chose to do this offline because this is a critical step, and we want to avoid an inconsistent state in AWS in case anything breaks during this procedure. Having this all file-based means it's easier to reason about what's happening and write test cases to ensure the logic is sound.

![Generate Target State](/img/2024/05/idc_migration_generate_target_state.png)

This merged permission structure is now our target state for the real environment. Next, we use another script to basically _diff_ the export of the current environment and the target state. This creates a list of changes needed to get from the current environment state to the target environment state. Essentially, this is a list of assignments to add or remove. Our changeset is stored as a file as well to enable easy verification and for the same consistency considerations.

![Generate Changeset](/img/2024/05/idc_migration_generate_changeset.png)

Subsequently, it's time to apply these changes to the real AWS environment, which another script does. It's written so that it can be stopped and restarted at any point, i.e., it can handle changes that have already been applied. We can restart the process if something goes wrong and we somehow lose connection.

After the changes have been applied, our system should be in the desired state. To double-check, we create another export and _diff_ the target state with the export - there shouldn't be any changes. Now, users should be able to log in using their new identities from the new Entra ID and have all the same permissions that they used to have before.

![Post Migration, pre Cleanup](/img/2024/05/idc_migration_pre_cleanup.png)

### Clean up

Switching the Identity Provider in the Identity Center leaves us with a long list of orphaned users and groups, as the SCIM-provided users and groups from the old Entra ID aren't deleted automatically. To avoid confusion, we're using a script to delete the users and groups that deletes users and groups whose external issuer (a reference to Entra ID) points to the old Entra ID.

There is another manual step here, in our case, that wasn't time-sensitive, though. Some tools used identifiers from the Identity Center to grant and restrict access to services like Quicksight. Since the underlying identifiers had to change, the policies had to be manually adjusted. This only affected a handful of users and wasn't worth automating.

## Dealing with Operational Problems

We tested the approach on a separate Identity Center environment first using the real Entra IDs involved in the live migration and designed the tools to be able to restart in case of problems. This means we were confident that the process would work, but nevertheless, we considered how things could go wrong.

Once we sever the connection to the old Entra ID IDP, and switch over to the new one, we basically hit a point of no return. Switching back to the old one is possible, but we have no guarantee that all identifiers remain identical, and if we were forced to do that, we'd be in a position where our tested tools had failed in a way that we hadn't foreseen and couldn't recover from. In that scenario, it's unlikely that they'd help us recreate the mapping.

This means we're at a point where we first focus on damage control, which means ensuring Ops personnel is able to access the production accounts to deal with any problems that may arise. In that scenario, we'd switch our Identity Center over to the internal Identity Store, create users for the Ops personnel, and manually assign them to the correct permission set.

While not an ideal situation, it would allow us to fix any issues with the scripts, APIs, or permissions and finish the migration.

## Non-obvious things that are important

When applying the changes, we need to ensure that new assignments are created before we delete any assignment. Otherwise, we could end up in a situation where an AWS account no longer has any assignment to a given permission set and the underlying IAM role is deleted. Later, it would be recreated but with a different identifier, which may break policies.

## Summary

This approach reduces both the downtime and risk of things going wrong while achieving the goal of migrating the identity providers. Additionally it allows us to verify the crucial mapping operations using automated tests.

If you're in a similar situation, [get in touch](https://www.tecracer.com/en/contact/), we're happy to support you in your migration.

---

Photo by [Julia Craice](https://unsplash.com/@jcraice) on [Unsplash](https://unsplash.com/photos/white-bird-faCwTallTC0) (Migrating Spoonbills)