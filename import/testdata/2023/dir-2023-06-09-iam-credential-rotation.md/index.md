---
title: "Advanced Credential Rotation for IAM Users with a Grace Period"
author: "Maurice Borgmeier"
date: 2023-06-09
toc: false
draft: false
image: "img/2023/06/luca-bravo-YoelVcKWmws-unsplash.jpg"
thumbnail: "img/2023/06/luca-bravo-YoelVcKWmws-unsplash.jpg"
categories: ["aws"]
tags:
  ["level-300", "iam", "python", "terraform", "scheduler", "well-architected"]
summary: |
  Rotating credentials with grace can be challenging when the underlying service doesn't support scheduled deletion. Today I will show you how to implement access key rotation for an IAM user while supporting a grace period where both the new and old credentials are valid.
---

The credentials attached to an IAM user are long-lived, which means a combination of access key id and secret access key doesn't expire by default. That makes using these kinds of credentials more risky than the short-term credentials offered by the security token service. The unfortunate reality is that some tools, e.g., Tableau, support only support long-term credentials to interface with AWS. While we generally try to use short-term credentials whenever possible, this is one of those cases where it isn't feasible.

That means we do the next best thing and find a way to artificially limit how long these credentials can be used by rotating them. The secret manager's secret rotation feature offers an approach to achieve this by implementing a Lambda function. Unfortunately, on its own, this has a few limitations. The lifecycle of a secret doesn't allow for a natural way to implement a grace period where both the old and new credentials are valid for a specific time.

If you have a manual process to exchange the credentials in the system relying on them, it's desirable to have a window where this exchange is made instead of a hard cut when new credentials are rolled out. Naturally, manual intervention should be the last resort here, but sometimes it's necessary.

Let's first explore how the secrets manager supports credential rotation and why this is not sufficient to solve our grace-period problem before moving on to the solution to our conundrum.

The basis for the secret rotation in the secrets manager is the following labels or version stages that can be applied to versions of a secret:

- `AWSCURRENT` is the currently active secret version; it will be fetched by default if you specify nothing else.
- `AWSPENDING` is a label for the new secret before it's validated and then transitioned to `AWSCURRENT`.
- `AWSPREVIOUS` is the label for the old secret after the `AWSCURRENT` label transitions to the new version.

The rotation process consists of four steps that manage these labels. A rotation is triggered according to a rotation schedule that you define. You must handle these [four steps](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html) in your Lambda function during a rotation:

1. `createSecret` - The first step is intended to generate new credentials and create a secret value with the `AWSPENDING` label
2. `setSecret` is the next step and should use the credentials from `AWSPENDING` to update the service that these credentials are for with the new value
3. `testSecret` can be used to validate that the credentials in `AWSPENDING` can be used to authenticate to the service
4. `finishSecret` is the last step and is used to transition the `AWSCURRENT` label to the `AWSPENDING` version so any future `GetSecretValue` [API](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html) call will return the updated credentials

This works well for situations where only one set of credentials can be active at any point in time. There's only one problem here. There is no built-in hook in the lifecycle to decommission or deactivate the old credentials after a grace period, so we have to build our own. Fortunately, the architecture for that isn't too complicated. We extend the four-step lifecycle with a fifth `deletePreviousSecret` step implemented through custom logic. A self-contained example implementation in Terraform and Python is [available here](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/terraform-graceful-secret-rotation). Let's go through each step in the sequence and see what's happening.

![Credential Rotation with Grace Periode - Architecture](/img/2023/06/credential_rotation_architecture.png)

During the `createSecret` stage, we use the `CreateAccessKey` [API](https://docs.aws.amazon.com/IAM/latest/APIReference/API_CreateAccessKey.html) to generate a new access key for the technical user and store it in the secrets manager using the `AWSPENDING` stage. In the next `setSecret` step we don't have much to do, but it takes a few seconds for new credentials to propagate through the whole system and become valid, so we just sleep for 10 seconds. Step 3 is `testSecret`, and here we read the `AWSPENDING` stage of the secret, and using those credentials, use the `GetCallerIdentity` [API](https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html) to check that they're valid credentials for the technical user.

The fourth step is where we deviate from the usual script. First, we update the secret so version with `AWSPENDING` is transitioned to `AWSCURRENT`. Next, we use the Event Bridge Scheduler [announced in November 2022](https://aws.amazon.com/blogs/compute/introducing-amazon-eventbridge-scheduler/) to schedule a future event for the Credential Updater Lambda, which will be the custom `deletePreviousSecret` step. This is set to occur at the end of the grace period (configurable).

```python
def finish_secret(service_client, arn, token, lambda_arn):

	# <transition AWSPENDING to AWSCURRENT>

    # Calculate when to delete the old credentials
    delete_in_n_minutes = int(os.environ.get(ENV_DELETE_OLD_AFTER_N_MINUTES, "5"))
    deletion_timestamp = (
        datetime.now() + timedelta(minutes=delete_in_n_minutes)
    ).isoformat(timespec="seconds")
    username = os.environ.get(ENV_IAM_USERNAME)

    # Schedule the deletion of the old credentials
    scheduler_client = boto3.client("scheduler")
    scheduler_client.create_schedule(
        Name=f"DeletePreviousAKFor{username}",
        ScheduleExpression=f"at({deletion_timestamp})",
        FlexibleTimeWindow={"Mode": "OFF"},
        Target={
            "Arn": lambda_arn,
            "RoleArn": os.environ.get(ENV_SCHEDULER_ROLE_ARN),
            # This is the event for step 5
            "Input": json.dumps(
                {
                    "Step": "deletePreviousSecret",
                    "ClientRequestToken": "not_relevant",
                    "SecretId": arn,
                }
            ),
        },
    )
```

When the Lambda function is invoked for the fifth time from the event bridge scheduler with the canned event we set up earlier, we request the old version of the secret with the `AWSPREVIOUS` label, extract the access key id and use the `deleteAccessKey` API to delete it permanently. To clean up after ourselves, we deleted the schedule that triggered the Lambda as well since that isn't removed on its own and would cause a naming conflict during the next rotation.

```python
def delete_previous_secret(secretsmanager_client, arn):

    # Get the previous access key id
    secret_response = secretsmanager_client.get_secret_value(
        SecretId=arn, VersionStage="AWSPREVIOUS"
    )
    creds = json.loads(secret_response["SecretString"])
    access_key_id = creds["access_key_id"]

    # Delete the old access key
    iam_client = boto3.client("iam")
    username = os.environ.get(ENV_IAM_USERNAME)
    iam_client.delete_access_key(
        UserName=username,
        AccessKeyId=access_key_id,
    )

    # Delete the schedule that triggered us.
    scheduler_client = boto3.client("scheduler")
    scheduler_client.delete_schedule(
        Name=f"DeletePreviousAKFor{username}",
    )
```

I'm only showing the excerpts from the code that implement the grace period; the whole solution is available as Terraform and Python in [this GitHub repository](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/terraform-graceful-secret-rotation). Let's talk about that. If you deploy the solution as outlined in the repository, it will create a secret called `technical-user-credentials` for you. Initially, that secret is empty, and the IAM user won't have any credentials. To change that, navigate to the secret in the AWS console, click on **Rotate secret immediately**, and confirm with **Rotate** in the dialog.

![Screenshot - Rotate Now](/img/2023/06/rotate_now_screenshot.png)

By default, the grace period is set to 10 minutes, so for the next 10 minutes, you'll see two sets of credentials attached to the IAM user, which you can request using the CLI:

```shell
# Current version
aws secretsmanager get-secret-value --secret-id \
technical-user-credentials --version-stage AWSCURRENT

# Previous version
aws secretsmanager get-secret-value --secret-id \
technical-user-credentials --version-stage AWSPREVIOUS
```

The old credentials will continue to function for 10 minutes, upon which they will be deleted. However, these credentials won't allow you to do much beyond calling `aws sts get-caller-identity`, as I haven't attached any permissions/policies to the technical user. After the 10 minutes are over, you'll still be able to request the old credentials from the secrets manager, but they won't do anything anymore.

One limitation is that the current setup is only built to support one IAM-User, and the permissions are very restricted to precisely that user, which is what we need for our use case. It could be extended to support more IAM-Users by moving the username to a tag at the secret and taking it from there while extending the permissions.

The permissions are a bit messier than usual, as two roles are involved, and a circular dependency between the two roles and the Lambda function had to be avoided through a managed policy. The Lambda function uses one role and has permission to update the secret, create and delete credentials for the IAM user and create and delete event bridge schedules with a specific name prefix. Additionally, it has permission to pass the other role, which the event bridge schedule uses to invoke the Lambda function. It's a bit tricky, but if you check out the code, you'll get the hang of it.

To summarize, I showed you how you can implement grace periods for IAM-User credentials in a secret rotation powered by the secrets manager. This isn't limited to IAM-Users, though. You could conceivably change that part out for any system supporting multiple authentication tokens for the same user, e.g., most API keys.

Thank you for your time, and hopefully, you learned something new!

&mdash; Maurice

---

Title Photo by [Luca Bravo](https://unsplash.com/@lucabravot) on [Unsplash](https://unsplash.com/photos/YoelVcKWmws)
