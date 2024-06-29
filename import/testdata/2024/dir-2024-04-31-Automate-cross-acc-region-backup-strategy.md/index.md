---
title: "Automating Cross-Account / Cross-Region Backups with AWS Backup in AWS Organizations"
author: "Dr Johann Nuck"
date: 2024-04-26
toc: true
draft: false
image: "img/2024/04/NUCK/AWS_Backup_Header_Image.jpg"
thumbnail: "img/2024/04/NUCK/AWS_Backup_Header_Image.jpg"
categories: ["aws"]
summary: |
  In this blog post we'll dive deep into AWS Backup.
  We cover how the service works, how to set it up and focus on how it interacts with different AWS resources.
  It's crucial to understand which features are supported for different services such as EBS or S3 to understand how to protect your environment.
  Additionally we look into Cross-Region and Cross-Account backup and restore options in the context of an AWS Organization.
tags:
  ["aws-backup", "backup", "level-300"]
---
  After deploying an infrastructure in AWS, the work isn't done yet. Reliable failure management has to be implemented to protect the different resources from attacks, hardware failure and oversight errors. Here we discuss the AWS Backup service with its possibilities and take a deep dive into its peculiarities.




## Introduction

Having a solid backup- and recovery strategy is fundamental when working with data. Therefore, it is a part of the Well-Architected Framework inside the [Reliability Pillar](https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/back-up-data.html).

Environments can be different, but backup strategies follow similar patterns. This blog post will not discuss these, as this is already done elsewhere ([1](https://docs.aws.amazon.com/prescriptive-guidance/latest/security-best-practices/strategy.html), [2](https://docs.aws.amazon.com/prescriptive-guidance/latest/backup-recovery/welcome.html)). Here we want to focus on the service [AWS Backup](https://aws.amazon.com/backup) which automates the backup process via [backup plans](https://docs.aws.amazon.com/aws-backup/latest/devguide/about-backup-plans.html) and can even be employed to [monitor](https://docs.aws.amazon.com/aws-backup/latest/devguide/manage-cross-account.html#enable-xcross-monitoring) the process from the management account across various accounts inside the organization.

There are already other blog posts about this service covering the [centralization](https://aws.amazon.com/de/blogs/storage/automate-centralized-backup-at-scale-across-aws-services-using-aws-backup/) and a specific case for [Amazon RDS and Aurora DBs](https://aws.amazon.com/de/blogs/database/automate-cross-account-backups-of-amazon-rds-and-amazon-aurora-databases-with-aws-backup/). So why write another blog about it? When having a look at the [feature availability by resource](https://docs.aws.amazon.com/aws-backup/latest/devguide/backup-feature-availability.html#features-by-resource) of the service (AWS Backup Developer Guide v14.03.2024), it is shown that not all resources have a "✓" in the [Full AWS Backup management](https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management) column. This missing check mark for most of the resources is crucial as the handling of the backups will be different. Furthermore, some items are marked with a (3) which implies some complications described in the 2nd [linked blog post](https://aws.amazon.com/de/blogs/database/automate-cross-account-backups-of-amazon-rds-and-amazon-aurora-databases-with-aws-backup/). 

The following blog post provides an overview of how to handle the different resources for backup and recovery. Common pitfalls and error messages will be discussed. Another aspect is to shine a light on the involved IAM roles, the IAM resource-based policies for the backup vaults and the resource policies for the AWS KMS encryption keys.

## Highlevel Overview of AWS Backup

The [AWS Backup Service](https://aws.amazon.com/backup/) is a fully managed solution to create backups (= recovery points) from resources on your behalf via *Backup Plans*. These plans consist of *Backup Rules* and *Resource Assignments*. The rules define the backup frequency, the backup vault where the recovery point is stored and an optional *Copy to Destination* option. This copy option can be chosen in case a recovery point shall be transferred into a backup vault of another AWS Account and/or AWS Region as well. The resource assignment (RA) can be done by resource type and/or tags.
In order to process the backup (or restore) task, the AWS Backup service requires specific permissions. These can be granted here by choosing an IAM role or the Default IAM role. This IAM role will be assumed by AWS Backup when creating and managing recovery points. Besides this IAM role, also another IAM role is involved in the process. The two involved IAM roles:

- The *AWS AWSBackupDefaultServiceRole* (or *BackupandRestore* IAM role)
- The *AWSServiceRoleForBackup* IAM role

will be discussed below. The recovery points are stored in a Backup Vault and encrypted with an AWS KMS encryption key. The recovery points can be restored directly from the backup vault or copied into another vault and restored afterward. This other vault can be in the same account/region and/or another account/another region. 

## Solution Overview

The following backup- and restore solution takes the [feature availability by resources](https://docs.aws.amazon.com/aws-backup/latest/devguide/backup-feature-availability.html#features-by-resource) into account. Each of the three resources in [*Table 1*](#aws_backup_table_summary)
 represents a group of resources to point out the different behavior of the AWS Backup service when creating and managing the recovery points within and 
across AWS accounts/ AWS regions.

<table id="aws_backup_table_summary" style="border:1px solid black;margin-left:auto;margin-right:auto;">
    <caption style="caption-side:bottom; text-align:left;"><em>Table 1: Summary of the functionality of AWS Backup when handling different resources. S3, EBS and RDS are employed as an example for the different resources.</em></caption>
<thead>
  <tr>
    <th class="tg-0pky">Resource</th>
    <th class="tg-0pky">Functionality</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="tg-0pky">S3</td>
    <td class="tg-dvpl"><a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management" target=”_blank”>Full AWS Backup management</a></td>
  </tr>  
</td> 
  <tr>
    <td class="tg-0pky">EBS</td>
    <td class="tg-dvpl"><a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/cross-region-backup.html" target=”_blank”>Cross-Region</a>/<a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/create-cross-account-backup.html" target=”_blank”>Account backup</a></td>
  </tr>
    <tr>
    <td class="tg-0pky">RDS</td>
    <td class="tg-dvpl">No single copy option Cross-Region/Account is supported</td>
  </tr>  
  </tbody>
</table>



### Backup- and Restore Solution Diagram
 
 [*Fig. 1*](#overview-backup) and [*Fig. 2*](#overview-restore) illustrate the backup- and restore solution discussed in this blog post.

<a id="overview-backup" href="/img/2024/04/NUCK/AWS_Backup_Backup_Solutions_Overview.png" target="_blank">
  <img src="/img/2024/04/NUCK/AWS_Backup_Backup_Solutions_Overview.png" alt="Overview of the backup solution">  
</a>
<p style="text-align:left; width:100%; font-style: italic;">
Fig. 1: Overview of the backup solution including three different accounts: The management account to roll out the backup plans (<span style="font-size: 120%">&#10122;</span>) and monitor cross-account/cross-region jobs, the member account, where the resources in question are hosted (<span style="font-size: 120%">&#127312;</span>) and the backup account to assure cross-account/cross-region backups (<span style="font-size: 120%">&#127313;</span>).
</p>
<br>
<a id="overview-restore" href="/img/2024/04/NUCK/AWS_Backup_Restore_Solutions_Overview.png" target="_blank">
  <img src="/img/2024/04/NUCK/AWS_Backup_Restore_Solutions_Overview.png" alt="Overview of the restore solution">  
</a>
<p  style="text-align:left; width:100%; font-style: italic;">
Fig. 2: Overview of the restore solution. Backups can be restored on-demand locally from the member backup vault (<span style="font-size: 120%">&#10122;</span>), restored inside the backup account environment (<span style="font-size: 120%">&#10123;</span>) or copied from here back to the member account (<span style="font-size: 120%">&#10124;</span>) and restored from there. Certain resources (eg. RDS) have to take a detour over the region eu-central-1 inside the backup account (<span style="font-size: 120%">&#10125;</span>).
</p>

### Description

A cross-account and cross-region solution consists minimum of three different accounts, all members of the same organization:
- `Management Account`
- `Member Account`
- `Backup Account`

#### Management Account

After opting in for the *Cross-account management* under *AWS Backup* &rarr; *Settings*, this account is employed to deliver the *backup policies* (= backup plans) to the attached member accounts ( [*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#10122;</span>) and to monitor the *Backup-, Restore and Copy jobs*.

<div class="sm_items" id="AWS_Backup_enable">
<figure>
    <img src="/img/2024/04/NUCK/AWS_Backup_enable.png" alt="ChatGPT Sample Question and Answer" width="170%" style="vertical-align:middle">
    <figcaption style="text-align:left">Fig. 3: Opting in for the Cross-account management under the service AWS Backup in the management account. Afterward, Cross-account monitoring is available here (see arrow).</figcaption>
</figure>
</div>

Forgetting to opt-in for the cross-region option ([*Fig. 3*](#AWS_Backup_enable)) while trying to copy a recovery point 
into another account would result in the following error:

```
Copy job failed. Cross-account copy feature is not enabled for the current organization.
```

In the In this solution, three different backup plans are created, shown in [*Table 2*](#Backup_Plans), to discuss the 
behavior of the AWS Backup service when creating recovery points in a local vault (within the 
`Member Account`) and copying them into another AWS account (`Backup Account`). [*Fig. 4*](#BackupPlan) shows how to create a backup plan via the console.

<div class="sm_items">
<figure>
    <img id="BackupPlan" src="/img/2024/04/NUCK/AWS_Backup_BackupPlan.png" alt="Creating a Backup Plan" width="120%" style="vertical-align:middle">
    <figcaption style="text-align:left">Fig. 4: Example for creating a backup plan via the console. Highlighted is the <b>Copy to destination</b> option. With this option, it is possible to copy the recovery point into another vault.</figcaption>
</figure>
</div>

##### Backup Plan: `Simple`

This plan creates recovery points inside the `Member Account` in the same region where the resources are stored.

##### Backup Plan: Cross Region (`CR`)

This plan creates recovery points inside the `Member Account` in the same region and a copy of these recovery points in the `Backup Account` in another region with the feature *Copy to destination* ([*Fig. 4*](#BackupPlan)). It can only be additionally activated: a local recovery point is mandatory.

##### Backup Plan: Cross Region-special (`CR-special`)

This plan creates recovery points inside the `Member Account` in the same region and a copy of these recovery points in the `Backup Account` in the same region. This is necessary as some resources (AWS RDS Service, AWS Aurora Service, AWS DocumentDB Service, and AWS Neptune Service) can not be copied cross-Region AND cross-account with a single copy action. Afterward, the recovery point is copied into the final region. This is done by a combination of AWS Eventbridge, AWS Lambda and AWS SQS but not further discussed here. More information on how 
to implement it in a similar way , can be found [here](https://aws.amazon.com/de/blogs/database/automate-cross-account-backups-of-amazon-rds-and-amazon-aurora-databases-with-aws-backup/).

Trying to copy it directly produces the following error:

```
Copy job from eu-north-1 to eu-central-1 cannot be initiated for Aurora resources. Feature is not supported for provided resource type.
```
This is the same case when trying the copy action for the restoring process.

The backup plans are called *global backup plans* and can only be edited and deleted from the management account.

### Member Account 

This account (with the pseudo account number `111111111111`) represents a member account where the resources are hosted which shall be backed up. From the management account, the global backup plans are deployed into this account. The resource assignment is done in this solution via tags: `backup - simple`, `backup - cross-region` and `backup - cross-region-special`. For each plan a representative resource is created. A summary can be found in [*Table 2*](#Backup_Plans):


<table class="tg" id="Backup_Plans">
    <caption style="caption-side:bottom; text-align:left"><em>Table 2: Backup plans to investigate the different functionalities of AWS Backup.</em></caption>
<thead>
  <tr>
    <th class="tg-0pky">Resource</th>
    <th class="tg-0pky">Functionality</th>
    <th class="tg-0pky">Backup Plan</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="tg-0pky">S3</td>
    <td class="tg-dvpl"><a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management" target=”_blank”>Full AWS Backup management</a></td>
    <td class="tg-0pky">Cross-Region (CR)</td>
  </tr>  
</td> 
  <tr>
    <td class="tg-0pky">EBS</td>
    <td class="tg-dvpl"><a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/cross-region-backup.html" target=”_blank”>Cross-Region</a>/<a href="https://docs.aws.amazon.com/aws-backup/latest/devguide/create-cross-account-backup.html" target=”_blank”>Account backup</a></td>
    <td class="tg-0pky">Cross-Region (CR)</td>
  </tr>
    <tr>
    <td class="tg-0pky">RDS</td>
    <td class="tg-dvpl">No single copy option Cross-Region/Account is supported</td>
    <td class="tg-0pky">Cross-Region-Special</td>
  </tr>
    </tr>
    <tr>
    <td class="tg-0pky">S3, EBS, RDS</td>
    <td class="tg-dvpl">see above respectively</td>
    <td class="tg-0pky">Simple</td>
  </tr>
  </tbody>
</table>
<link href="">

The `Member Account` needs a backup vault where the recovery point is stored. Depending on the backup plan, a copy of this recovery point can also be stored in another account/region. Here this option is utilized for the backup plans `CR` and `CR-special`. The backup plan `Simple` creates a recovery point only locally (=inside the `Member Account`).

### Backup Account

This is the backup account where recovery points are stored cross-account and cross-region. We assign the pseudo account number `222222222222`. Recovery points shall be stored additionally in a different region (`eu-north-1`) than the original region (`eu-central-1`). As not all resources can be copied cross-account and cross-region, two backup vaults have to be created in each region. The following resources [have to](#backup-plan-cross-region-special-cr-special) take a detour through the region `eu-central-1`:

- AWS RDS Service
- AWS Aurora Service
- AWS DocumentDB Service
- AWS Neptune Service

For these resources, the backup plan `CR-special` is created. The recovery point is first copied into the `Backup account` in the same region (`eu-central-1`) and then copied to the final region (`eu-north-1`). This behavior is automated with a combination of AWS EventBridge, AWS Lambda and AWS SQS. This blog will not explain the detailed solution, as it is already similarly done [elsewhere](https://aws.amazon.com/de/blogs/database/automate-cross-account-backups-of-amazon-rds-and-amazon-aurora-databases-with-aws-backup/).

## Detailed Description

### Necessary AWS IAM Roles

When we create a backup plan and assign resources or create an on-demand backup, we also have to define an IAM role. This has to be done also for on-demand restoration. This IAM role will be assumed by the AWS Backup service to perform the backup/restoration.

<div class="sm_items" id="IAM_Role">
<figure>
    <img id="on-demand" src="/img/2024/04/NUCK/AWS_Backup_on-demand.png" alt="" width="170%" style="vertical-align:middle">
    <figcaption>Fig. 5: Choosing a custom IAM role when creating an on-demand backup.</figcaption>
</figure>
</div>

AWS lets you choose the default IAM role [AWSBackupDefaultServiceRole](https://docs.aws.amazon.com/aws-backup/latest/devguide/iam-service-roles.html#default-service-roles) or one can be created ([*Fig. 5*](#IAM_Role)). The default AWS IAM role includes the following permission policies:
- [AWSBackupServiceRolePolicyForBackup](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AWSBackupServiceRolePolicyForBackup.html)
- [AWSBackupServiceRolePolicyForRestores](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AWSBackupServiceRolePolicyForRestores.html)

This IAM role [cannot handle](https://docs.aws.amazon.com/aws-backup/latest/devguide/s3-backups.html#one-time-permissions-setup) backing up and restoring S3 resources. The following error message will appear if you try to choose this role for S3:

```
IAM Role arn:aws:iam::111111111111:role/service-role/AWSBackupDefaultServiceRole does not have sufficient permissions to execute the backup.
```

To be able to backup S3 buckets, a custom AWS IAM role `BackupandRestore` is created manually, including the two previous permission policies and enhanced with the policies:
- [AWSBackupServiceRolePolicyForS3Backup](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AWSBackupServiceRolePolicyForS3Backup.html)
- [AWSBackupServiceRolePolicyForS3Restore](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AWSBackupServiceRolePolicyForS3Restore.html)

This is not the only IAM role involved in the backup- and restore process. AWS has created an [AWS service-linked role](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_terms-and-concepts.html#iam-term-service-linked-role) for AWS Backup. This IAM role can only be assumed by the specific service due to its [trust policy](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_terms-and-concepts.html#term_trust-policy).<br>
The [AWSServiceRoleForBackup](https://docs.aws.amazon.com/aws-backup/latest/devguide/using-service-linked-roles-AWSServiceRoleForBackup.html) provides the AWS Backup with parts of the permission necessary to perform its service.

{{% notice note %}}
The permissions necessary to work with the AWS Backup service are separated into two roles:
The service-linked IAM role `AWSServiceRoleForBackup` and the service IAM role `AWSBackupDefaultServiceRole` (or our custom IAM role `BackupandRestore`).
{{% /notice %}}

If the AWS Backup service has never been used in an account, these IAM roles don't exist and have to be created first by eg. starting an on-demand backup job. Choosing here the *Default IAM role*, the error:
```
IAM Role arn:aws:iam::111111111111:role/service-role/AWSBackupDefaultServiceRole does not have sufficient permissions to execute the backup.
```
will appear. This message is not correct as the IAM role just doesn't exist yet. But due to this action, the IAM role has been created in the background. After clicking on *Create on-demand backup*, the IAM role `AWSServiceRoleForBackup` will also be created.


### Encryption and Access

When a backup plan runs, it creates a recovery point from the assigned resources and stores them in a backup vault. These recovery points are encrypted. But how they are encrypted, depends on the resource itself and whether it stays in the same account and region or is copied somewhere else. We have to distinguish between resources supporting full 
AWS Backup management like S3 and not fully supporting resources. 

We assume that all three representative resources in the `Member Account` in region `eu-central-1` are encrypted with a custom-managed **resource** key (CMRK) ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#127312;</span>). The three backup vaults in the different accounts and regions have a custom-managed **vault** key (CMVK) ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#127313;</span>). It is a good idea to encrypt the backup vaults with a custom-managed key instead of the aws-managed backup key (Alias: *aws/backup*). Otherwise copying the recovery point into another vault will fail for resources not supporting [full AWS Backup management](https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management) with an error:

```
Copy job failed because the destination Backup vault is encrypted with the default Backup service managed key. The contents of this vault cannot be copied. Only the contents of a Backup vault encrypted by an AWS KMS key may be copied.
```

The backup plan `Simple` creates only a recovery point for the resources in the same account and region ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#10123;</span>). Running this backup plan, the following behavior is observed:


{{% notice note %}}
Resources which support full AWS Backup management like S3 are decrypted and re-encrypted with the CMVK ([*Fig.1* <span style="font-size: 120%">&#127315;</span>](#overview-backup)). This feature is called [Independent encryption](https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management).<br> Not fully supported resources (AWS EBS Service, AWS RDS Service) keep their CMRK ([*Fig.1* <span style="font-size: 120%">&#127315;</span>](#overview-backup)).
{{% /notice %}}


When the *Copy to destination* option is set, as shown in <a href="#BackupPlan">*Fig. 4*</a>, the situation is different. The backup plan *CR* is designed to first create a recovery point in the same account/region and copy then the recovery point into a backup vault in another account (`Backup account`) and another region (`eu-north-1`) ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#10124;</span>). The backup plan `CR-special` is similar but copies the recovery point first into the vault of another account in the same region as this is [mandatory](#backup-plan-cross-region-special-cr-special) for RDS ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#10125;</span>). By design of the Backup service, it is not possible to copy the recovery point directly info a foreign vault without creating it prior locally (= creating a recovery point in the same AWS account and region where the resource is located). <br>
For the **copied** recovery points, the following behavior is observed:

{{% notice note %}}
All resources are decrypted and re-encrypted with the specific CMVK of the destination vault ( [*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#127316;</span>).
{{% /notice %}}
Access and usage of resources are controlled by permissions via policies. Not only do the AWS KMS keys for the backup resources and recovery points have policies attached, but also the vaults where the recovery point are stored. This circumstance leads to some questions to find out the correct policies for the different keys and vaults:

- Who is allowed to copy into a specific backup vault?
- Who does the decryption / re-encryption?


### Who is allowed to copy into a specific backup vault

Storing recovery points from resources in the same account where the backup vault is located doesn't need any access policy.Access for the AWS Backup serivce to the backup vault has to be granted for cross-account backups. Missing this permission will result in the error:
```
Access Denied trying to call AWS Backup service.
```

When copying recovery points into a backup vault of another account, a [common practice](https://docs.aws.amazon.com/aws-backup/latest/devguide/create-cross-account-backup.html) is to allow all members of the organization to copy recovery points into a vault with the following *Vault Access Policy*:

```JSON
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":"*",
      "Action":"backup:CopyIntoBackupVault",
      "Resource":"*",
      "Condition":{
        "StringEquals":{
          "aws:PrincipalOrgID":[
            "o-a1b2c3d4e5"
          ]
        }
      }
    }
  ]
}
```

Here a `Condition` is set where the `PrincipalOrgID` has to match the organization ID "o-a1b2c3d4e5". This policy statement can be attached to all backup vaults ([*Fig. 1*](#overview-backup): <span style="font-size: 120%">&#127317;</span>). This policy is also a predefined option when adding a permission to the backup vault in the console.

Of course, further restrictions can be implemented by setting a principal eg. `root` or IAM role `myRole` of a specific account to follow the best practice of the [least privilege principle](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html):

```json
"Principal":{
  "AWS":[
    "arn:aws:iam::111111111111:root",
    "arn:aws:iam::222222222222:role/myRole"
  ]
}
```


### Who does the decryption / re-encryption

In order to see who is responsible for these actions, it is a good idea to look into the encryption key polices. Here the *Principal* has to be defined. <br>
For the creation of the recovery point inside the **same account** (`Member account`), the [default key policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-default.html#key-policy-default-allow-root-enable-iam) is enough: 

```JSON
{
  "Sid": "Enable IAM User Permissions",
  "Effect": "Allow",
  "Principal": {
    "AWS": "arn:aws:iam::111111111111:root"
   },
  "Action": "kms:*",
  "Resource": "*"
}
```

The principal here is set to the [AWS account principal](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-accounts). This also includes the service AWS Backup.
In case the recovery point is furthermore copied into **another account** (`Backup account`), it has to be decrypted first, and then re-encrypted with the new CMVK from the new vault. Via [AWS Cloudtrail](https://aws.amazon.com/cloudtrail/) it is possible find out that the *AWSServiceRoleForBackup* from the **destination account** wants to access the decryption key (here CMRK for AWS EBS Service and AWS RDS Service, CMVK for AWS S3 Service). It is important to know that it is **not** the IAM role we <a href="#on-demand">choose</a> for e.g. our on-demand backup plan (*AWSBackupDefaultServiceRole* / *BackupandRestore*). More information about the task of this 
default role can be found [here](https://docs.aws.amazon.com/aws-backup/latest/devguide/iam-service-roles.html).

 Therefore the default key policy has to be modified in a way that cross-account access from the *AWSServiceRoleForBackup* role is possible. A summary of cross-account KMS key access can be found [here](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html). 
 The CreateGrant statement has to be added to the default key policy 
statement shown above. This will allow the AWSServiceRoleForBackup  role from the Backup 
account to use the KMS key in cryptographic operation – here to decrypt the encrypted recovery 
point. 

```json
{
    "Sid": "Allow attachment of persistent resources",
    "Effect": "Allow",
    "Principal": {
        "AWS": "arn:aws:iam::222222222222:role/aws-service-role/backup.amazonaws.com/AWSServiceRoleForBackup"
    },
    "Action": [
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant"
    ],
    "Resource": "*",
    "Condition": {
        "Bool": {
            "kms:GrantIsForAWSResource": "true"
        }
    }
}
```
Otherwise copying a recovery point into another account produces a resource-dependent (!) error eg.

- AWS RDS Service
  ```
  The source snapshot KMS key [arn:aws:kms:eu-central-1:111111111111:key/159f9as4-697q-f354-a448-example]
  does not exist, is not enabled or you do not have permissions to access it.
  ```

- AWS EBS Service
  ```
  Given key ID is not accessible
  ```

As shown in the policy, the principal is the *AWSServiceRoleForBackup* from the `Backup Account` (destination). In case we want to copy back our recovery point for restoring purpose from the `Backup Account` into the `Member Account`, the situation is inverse:  the *AWSServiceRoleForBackup* from the `Member Account` needs access to the CMVK from the `Backup Account`. Therefore the principal for the key policy of the vaults in the `Backup Account` is:

```json
    "Principal": {
        "AWS": "arn:aws:iam:::111111111111:role/aws-service-role/backup.amazonaws.com/AWSServiceRoleForBackup"
    }
```
Be aware that these roles have to already exist, otherwise, the following error will appear:

```
PutKeyPolicy request failed 
MalformedPolicyDocumentException - Policy contains a statement with one or more invalid principals.
```

### Restore

There are several options for restoring resources as we created various copies in different accounts and regions. Depending on the case of the incident, a fitting option can be picked.

#### Local resource in the `Member Account` is lost (deleted/corrupted etc.)

In this case, it is enough to restore the resource from the local backup vault ([*Fig. 2*](#overview-restore): <span style="font-size: 120%">&#10122;</span>). This can be done eg. via the AWS Backup console by choosing the recovery point and clicking *Restore*. Under *Restore role* choose either the default IAM role (*AWS AWSBackupDefaultServiceRole*) or the created *BackupandRestore* IAM role for S3. To simplify the restore process for S3, it is easier to create a new bucket for S3 instead of restoring it into the same buckets. This will avoid permission problems to write into an existing bucket.

Regarding encryption, a KMS key can be chosen freely for AWS S3 Service and AWS EBS Service. In the case of AWS RDS Service, it depends on the database engine utilized: eg. for MySQL the restored database is bound to its initial recovery point encryption key. The encryption key for AWS Aurora Service can be freely chosen during the restore process.

#### The region is not reachable or account access is lost

The previous option is not possible when either the complete region is [not reachable](https://aws.amazon.com/premiumsupport/technology/pes/) (eg. outage) or the `Member Account` has been [compromised](https://repost.aws/knowledge-center/potential-account-compromise). 
To cover this scenario, the more complex and cost-expensive CR / CR-special backup strategy has to be chosen. The resources are safely backed up cross-region and cross-account.

For this situation, there are several restoration options. It is possible to restore the recovery points locally in the `Backup Account` in the region `eu-north-1` in the discussed solution ([*Fig. 2*](#overview-restore): ➋). This is done the same way as described above. Depending on the type of incident, resources may be needed in the `Member Account` in region `eu-central-1`. Recovery points can be copied into other vaults and restored afterward from there ([*Fig. 2*](#overview-restore): ➌). The same rules are applied as for *Copy to destination* discussed previously:

- The backup vault has to be [accessible](#who-is-allowed-to-copy-into-a-specific-backup-vault) from the origin account
- The KMS encryption key has to be [accessible](#who-does-the-decryption--re-encryption) by the role *AWSServiceRoleForBackup* from the destination account
- Similar to the backup process, [certain resources](#backup-plan-cross-region-special-cr-special) (e.g. AWS RDS Service) have to take a detour (e.g. first cross-account copy and then a cross-region copy  ([*Fig. 2*](#overview-restore): ➍) 

## Key findings and summary

AWS Backup is a helpful service to support your backup- and restore strategies. It can be employed to create recovery points inside a backup vault in an AWS account and AWS region and/or copy them into other vaults in different AWS accounts/AWS regions.

Two IAM roles are involved in the backup- and restore process: The service-linked IAM role 
*AWSServiceRoleForBackup* and the service IAM role *AWSBackupDefaultServiceRole*. 
The behavior of the AWS Backup service for resources can be grouped into three categories sorted 
descending by number of features: 
- Full AWS Backup management support e.g. AWS S3 Service 
- Cross-Region/Account backup support e.g. AWS EBS Service 
- No single copy option Cross-Region/Account support e.g. AWS RDS Service

Recovery points from resources supporting full AWS Backup management are always encrypted with the customer-managed vault key from the backup vault. Recovery points from non-supporting full AWS Backup management resources keep their customer-managed resource key when created. But when these recovery points are copied into another backup vault, they will be re-encrypted with the customer-managed vault key from the destination vault. Therefore, the KMS keys of the recovery points have to be accessible by the role *AWSServiceRoleForBackup* from the destination account. 

For local backups, no vault access policy for the backup vault has to be defined. For cross-account copies, the destination vault access policy has to be modified in a way that access for `"Action":"backup:CopyIntoBackupVault"` is allowed. Some resources (e.g. AWS RDS Service) have to be handled specifically when copying recovery points cross-region and cross-country. They can only be copied in two actions eg.: 

1. cross-account + same-region 
2. same-region + cross-region

&mdash; Johann
