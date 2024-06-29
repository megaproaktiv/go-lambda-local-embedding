---
title: "How To Hybrid! - Secure AWS Environment Service Access"
author: "Marco Tesch"
date: 2022-02-09
draft: false
toc: true
image: "img/2021/09/hybrid-cloud-patch-management.png"
thumbnail: "img/2021/09/hybrid-cloud-patch-management.png"
categories: ["aws"]
tags: ["ssm", "credentials", "temporary", "hybrid", "linux", "on-premises", "ubuntu", "systems-manager", "aws", "activations", "advanced", "level-100", "level-200"]
summary: |
    As AWS Cloud adoption becomes more widespread throughout the industries, challenges arise how to govern IT resource usage and implement a coherent management for systems across on-premises and the AWS Cloud. This blog post gives insights in how the AWS offered Systems Manager capabilities could be extended from the cloud environment to your hybrid and on-premises environments.
---

Main goal of this part of the How To Hybrid series will be the security settings needed to provide secure access for your on-premises servers to AWS services, while using dynamic and temporary credentials instead of the creation of static credentials.

This second part of the How To Hybrid series will build on top of the first part, where we covered the topics of Hybrid Activations and AWS SSM Managed Instances. Please refer to [those sections]( /2021/09/how-to-hybrid-aws-systems-manager-patch-management.html#steps-to-hybrid-patch-management) if you have trouble activating your on-premises instances to AWS SSM.

## Lets obtain AWS Instance Profile like Credentials

To access AWS we always want to use temporary credentials, to get our on-premise servers to behave more like an EC2 Instance we can make use of the Amazon SSM Agent configuration. Therefore we take a look at the directory `/etc/amazon/ssm` within that directory we will locate the file `amazon-ssm-agent.json.template`, which is the configuration file for our Amazon SSM Agent. The most important properties we need to configure are shown below:

### Amazon SSM Agent Config Properties

```json
{
    "Profile":{
        "ShareCreds" : true,
        "ShareProfile" : "",
        "ForceUpdateCreds" : false,
        "KeyAutoRotateDays": 0
    },
    …
}
```

1.	Property `ShareCreds` is enabled by default, and we want to keep that setting active, to instruct the SSM Agent to obtain Credentials for our Managed Instance.
2.	Property `ShareProfile` is set to an empty `string` value and we want to change that value to our preferred profile name we want to use within our applications or AWS CLI calls.
3.	Property `ForceUpdateCreds` is disabled by default, and we can change it to true if we want the Amazon SSM Agent to overwrite our existing shared credentials file if it cannot be parsed, which most likely is related to an error within the file.
4.	Property `KeyAutoRotateDays` is the most important property to make our on-premises Managed Instance use temporary credentials. The default is set to `0` which instructs the Amazon SSM Agent to never rotate our obtained credentials which we want to change definitely! I advise to set this value to `1` and instruct the Amazon SSM Agent to rotate our credentials daily.

> Before we change any values make sure to copy the template file to the active configuration file, for example by using the following command from within the directory `/etc/amazon/ssm`

```shell 
sudo cp amazon-ssm-agent.json.template amazon-ssm-agent.json 
```

### Amazon SSM Agent Config Changes and Activation

I changed the settings on my test system to the following structure:

```json
{
    "Profile":{
        "ShareCreds" : true,
        "ShareProfile" : "az-user",
        "ForceUpdateCreds" : false,
        "KeyAutoRotateDays": 1
    },
    …
}
```

This allows the Amazon SSM Agent to obtain temporary credentials for my on-premise Managed Instance, rotate the credentials every day and store the credentials in the shared credentials file (located in `/root/.aws/credentials`).

The next step would be to restart the Amazon SSM Agent service to pick up our newly created configuration and obtain our temporary credentials. Just execute the following command to do so:

```shell
sudo systemctl restart amazon-ssm-agent
```

### Verify Temporary Credentials

After we followed the above mentioned steps we can find our temporary credentials located within the file `/root/.aws/crentials`, which is the shared credentials file used by our amazon-ssm-agent (as it is run as a service using the Linux root user). We can verify that everything worked out correctly by executing the following command:

```shell
sudo cat /root/.aws/credentials
```

> If you followed along with the same configuration, I used for the Amazon SSM Agent you should find a shell output like 

```shell
[az-user]
aws_access_key_id     = ASI…FWE
aws_secret_access_key = 3n/…V8C
aws_session_token     = IQo…co=
```

As you can see the credentials are stored in a profile which is the same as configured in the configuration file by setting the property `ShareProfile` to `az-user`. But whats next? – you might ask, let’s make these credentials usable for a user which is not the Linux root user in the first step.

## Usage of Temporary AWS Credentials as non-root user

In the first part of this blog post, we managed to obtain temporary credentials for our on-prem Managed Instance which are automatically rotated daily. Unfortunately, as our Amazon SSM Agent service is executed as the root user on our Linux server those credentials are stored in a file located in the root users home dir `/root/.aws/credentials`. The second part of this blog post should now describe a way to make those credentials usable to non-root users on our on-prem server to give us more flexibility by keeping the security and governance of our auto rotating AWS Credentials.

### Setting needed Linux permissions

We will begin by setting the needed Linux permissions to gain access to the shared credentials file which is managed by our Amazon SSM Agent. Therefore, we will give read access to the following directories and folders

> As I will not go into detail on Linux filesystem permissions, please refer to your preferred documentation on that regard to understand the implications on the configuration and steps performed during this part of the blog post

-	`/root`
-	`/root/.aws`
-	`/root/.aws/credentials`

```shell
sudo chmod a+r /root /root/.aws /root/.aws/credentials
```

As well as execute permissions for everyone on the Linux system on the following directories to allow for path traversal for the users accessing the credentials with absolute paths:

-	`/root`
-	`/root/.aws`

```shell
sudo chmod a+x /root /root/.aws
```
Now you should be able to access the shared credentials file without the need of `sudo`. Please verify by executing the following command and comparing the output to the verification output you received in the first part of this blog post:

```shell
cat /root/.aws/credentials
```

### Bring convenience to Temporary Credentials

Now we have a shared credentials file which is located in `/root/.aws/credentials`, is managed by our Amazon SSM Agent and can be accessed by non-root Linux users. The next step would be to get some convenience to our environment by linking the shared credentials file to our Linux users home directory for all our AWS tools to automatically pick up those credentials. To achieve that behavior, we simply create a symbolic link within our users home directory to the managed shared credentials file by executing the following command:

```shell
ln -s /root/.aws/credentials ~/.aws/credentials
```

Once we did that we should now be able to easily access the content of our shared credentials file by executing our cat command on our symbolic link like so:

```shell
cat ~/.aws/credentials
```

> You again can compare the results of this command with the previously executed commands to verify the correct setup and configuration. Depending on your settings on the IAM Role associated with the Hybrid Activation you might already see a different output as SSM already rotated your credentials.

### Testing AWS access using AWS CLI (optional)

This final part is fully optional, and we will go into more detail on use cases for the shared credentials in one of the following blog posts in the How To Hybrid series. Depending on your Hybrid Activations associated IAM Role and if you have the AWS CLI installed on your system you might be able to verify the access to the AWS API by executing the following command which lists your AWS SSM Associations:

```shell
aws ssm list-associations --profile az-user --region eu-central-1
```

You should now see a response from the AWS API presenting you with a JSON document and using your ShareProfile from the shared credentials file.

## Conclusion and outlook

This second blog post gave you an introduction in obtaining, managing, and using temporary AWS credentials on your on-premises instances, a feature which imitates the behavior of EC2 Instance Roles. Depending on your IAM Role session duration configuration you will get new credentials at least daily. This increases security by magnitudes compared to static credentials you store for technical/human users on your on-premises servers. One of the next blog posts within the How To Hybrid series will go into a more detailed use case where you could apply the described pattern to extend the capabilities of your Hybrid Environment to connect to the AWS Cloud and AWS Managed Services you have provisioned.
