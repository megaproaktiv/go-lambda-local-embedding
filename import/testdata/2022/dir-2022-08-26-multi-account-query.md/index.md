---
title: "Find all Lambda-Runtimes in all Accounts: Multi Account Query with steampipe and TASFKAS (the AWS service formerly known as SSO *)"
author: "Gernot Glawe"
date: 2022-08-26
draft: false
image: "img/2022/08/kraken.jpeg"
thumbnail: "img/2022/08/kraken.jpeg"
toc: true
keywords:
  - security
  - IAM Identity Center
  - steampipe
tags:
  - level-300
  - IAM Identity Center
  - security
  - well-architected

categories: [aws]
---

You have got some mails from AWS: [Action Required] AWS Lambda end of support for Node.js 12
[Action Required] AWS Lambda end of support for Python 3.6 [Solution Required] Search all Lambdas in multiple accounts.

[Solution Found] Steampipe with AWS multi-account support. Multi-account management is like managing all the arms of a Kraken. I will show you a fast and straightforward solution for this.
(\* the new offical name is IAM Identity Center, but I think TASFKAS would also fit 😉)

<!--more-->

## Architecture multi-account query solution

![overview](/img/2022/08/steampipe-multi-overview.png)

**(1)** In the _IAM Identity Center_, we define a _Permission Set_, which includes a _AWSPowerUserAccess_ managed policy.

**(2)** Your user is assigned to a AWS Managed Policy `AWSPowerUserAccess` in all accounts, which should be queried

**(3)** All accounts are referenced in AWS user profiles

**(4)** With `aws-sso-util` we login into SSO

**(5)** All profiles are configured for Steampipe

**(6)** Steampipe does multi-account queries with all AWS user profiles

## Step 1 - Configure

### Steampipe - query AWS with SQL

What is [Steampipe](https://steampipe.io/docs)? It is a solution to query your AWS account (and many more things) with SQL. The `AWS Compliance Mod` gives you a dashboard supporting "configuration, compliance and security controls or full compliance benchmarks", including the "AWS Foundational Security Best Practices". The data for the dashboard is queried with SQL.

The installation is straightforward and documented on the website: [install Steampipe ](https://steampipe.io/downloads).

For a Mac, the actions are:

```bash
brew tap turbot/tap
brew install steampipe
steampipe -v
```

The output for `steampipe -v` should be sth like:

```log
steampipe version 0.16.0
```

We need to install the AWS plugin:

```bash
steampipe plugin install aws
```

This translates AWS Resources to SQL data tables. After the general AWS access is provided by the AWS plugin, you may choose from different [mods](https://hub.steampipe.io/mods). For a quick overview, there is the [aws insights mod](https://hub.steampipe.io/mods). For this usecase we use another mod, the [AWS Compliance Mod](https://hub.steampipe.io/mods/turbot/aws_compliance).

### Steampipe Mod for AWS Compliance

You clone the repository and start steampipe within this directory. You can also reference the working directory with the parameter `--workspace-chdir`.
To do this, we save the current working directory in `MOD`:

```bash
git clone https://github.com/turbot/steampipe-mod-aws-compliance.git
cd steampipe-mod-aws-compliance
export MOD=`pwd`
```

The next step is to configure multi-account authentication.

### Configuring Multi-Account

You could just use static credentials. More elegant is the Single Sign on solution, because you do not use static credentials.

### Installing `aws-sso-util`

For SSO on the cli, we use this tool:

[GitHub - benkehoe/aws-sso-util: Smooth out the rough edges of AWS SSO (temporarily, until AWS makes it better).](https://github.com/benkehoe/aws-sso-util)

With this tool you login to AWS-SSO. Then you can use AWS user profiles to access all accounts, which are configured in the _IAM Identity Center_.

#### Aws user profiles

Locally on your laptop, you configure AWS user profiles matching the SSO profiles

```bash
vi ~/.aws/config
```

```yaml
[profile megaproaktiv_dev]
sso_start_url=https://${yourssoid}.awsapps.com/start/
sso_region=eu-central-1
sso_account_name=megaproaktiv_dev-dev
sso_account_id=${youraccountid}
sso_role_name=AWSPowerUserAccess
credential_process=aws-sso-util credential-process --profile  megaproaktiv_dev
sso_auto_populated=true
```

This is an example config for a profile.
You need the sso-id `${yourssoid}`, it looks like "d-1234567890", the name of the role, here "AWSPowerUserAccess" and the account number `${youraccountid}`, which is a 12 digit number.

Please note: the AWS config file does not support variables, you have to replace them with real world values.

Examples can be found on [github](https://github.com/megaproaktiv/aws-community-projects/tree/main/steampipe-multi-account-query).

#### Login

With this configured, you start with a sso login.

```bash
aws-sso-util login https://d-1234567890.awsapps.com/start/
```

Where `d-1234567890`is you sso-id.

![](/img/2022/08/approved.png)

In the browser you login with your sso credentials and approve the request.

With a simple command, you can check, whether the login and the configuration is ok:

```bash
aws sts get-caller-identity  --profile megaproaktiv_dev
{
    "UserId": "AROA3SHER36FCZHIFNNGX:john@doe.com",
    "Account": "123456789012",
    "Arn": "arn:aws:sts::123456789012:assumed-role/AWSReservedSSO_AWSReadOnlyAccess_e92dcd3dfcd065e3/john@doe.com"
}
```

You should see:

- The userid of a temporarily created user - a AccessKey beginning with "AROA"-, followed by your SSO email
- The number of the account referenced in the AWS user profile
- The arn of the role you have assumed

For an explanation of the unique ID prefixes like `AROA` or `AIDA`, see the [AWS Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html).

Check other profiles

```bash
aws sts get-caller-identity  --profile megaproaktiv_test
aws sts get-caller-identity  --profile megaproaktiv_prod
```

Congrats, authorization is done!

### Steampipe config multi Account

Now the aws authentication is done, next step is to tell steampipe to use all of these profiles:

The main configuration is in `~/.steampipe/config/aws.spc`

You define the general AWS connection and the aggregator for all accounts:

```hcl
connection "aws" {
  plugin = "aws"

  regions = ["eu-central-1"]
}

connection "megaproaktiv" {
 type        = "aggregator"
 plugin      = "aws"
 connections = ["megaproaktiv_*"]
}

```

And one entry per account:

```hcl

connection "megaproaktiv_prod" {
  plugin = "aws"
  regions = ["eu-central-1"]
  profile = "megaproaktiv-prod"
}

connection "megaproaktiv_test" {
  plugin = "aws"
  regions = ["eu-central-1"]
  profile = "megaproaktiv-test"
}


connection "megaproaktiv_dev" {
  plugin = "aws"
  regions = ["eu-central-1"]
  profile = "megaproaktiv-dev"
}
```

All this configuration has to be done only _once_. Then you can query!

## Step 2 - Query with dashboard

### Standard dashboard

Now start steampipe in the terminal where you did the sso-login:

```bash
steampipe dashboard --search-path-prefix megaproaktiv --workspace-chdir ${MOD}
```

This will start your standard browser:

![Dashboard Start](/img/2022/08/dashboard-start.png)

And give some logs in the terminal:

```log
[ Wait    ] Loading Workspace
[ Wait    ] Starting Dashboard Server
[ Message ] Workspace loaded
[ Message ] Initialization complete
[ Ready   ] Dashboard server started on 9194 and listening on local
[ Message ] Visit http://localhost:9194
[ Message ] Press Ctrl+C to exit
[ Wait    ] Dashboard execution started: aws_compliance.benchmark.foundational_security
```

Now click on "AWS Foundational Security Best Practices
ComplianceAWSBenchmark" _in the browser_ and wait for `Ready` _in the terminal_. This will only take seconds!

```log
[ Ready   ] Execution complete: aws_compliance.benchmark.foundational_security
```

After that you have to refresh the browser page once and you get the AWS Foundational Security Best Practices
overview:

![Foudational](/img/2022/08/foundational-report.png)

Scroll to lambda and klick "2 Lambda functions should use latest runtimes".

You get a full **multi-account** list of all your lambdas. The accounts are blurred in the picture because of security.

![Lambdas](/img/2022/08/lambda-report.png)

**(1)** The name of the control

**(2)** The name of the Lambda functions

**(3)** region and account number

Oh, it warns about "Python 3.9", thats wrong! Python 3.9 is a current Lambda runtime. That is because the foundational_security can't keep up with all these changes. No problem, just update the SQL!

### Change the query

In the `steampipe-mod-aws-compliance`directory, you see the query for the dashboard

```bash
vi query/lambda/lambda_function_use_latest_runtime.sql
```

```sql
select
  -- Required Columns
  arn as resource,
  case
    when package_type <> 'Zip' then 'skip'
    when runtime in ('nodejs14.x', 'nodejs12.x', 'nodejs10.x', 'python3.8', 'python3.7', 'python3.6', 'ruby2.5', 'ruby2.7', 'java11', 'java8', 'go1.x', 'dotnetcore2.1', 'dotnetcore3.1') then 'ok'
    else 'alarm'
  end as status,
  case
    when package_type <> 'Zip' then title || ' package type is ' || package_type || '.'
    when runtime in ('nodejs14.x', 'nodejs12.x', 'nodejs10.x', 'python3.8', 'python3.7', 'python3.6', 'ruby2.5', 'ruby2.7', 'java11', 'java8', 'go1.x', 'dotnetcore2.1', 'dotnetcore3.1') then title || ' uses latest runtime - ' || runtime || '.'
    else title || ' uses ' || runtime || ' which is not the latest version.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_lambda_function;
```

Now I am saying, only the newest [runtimes](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html) of my favourite languages are allowed. See line 6:

```sql
    when runtime in ('nodejs14.x', 'nodejs12.x', 'nodejs10.x', 'python3.8', 'python3.7', 'python3.6', 'ruby2.5', 'ruby2.7', 'java11', 'java8', 'go1.x', 'dotnetcore2.1', 'dotnetcore3.1') then 'ok'
```

Change to

```sql
    when runtime in ('nodejs16.x',  'python3.9',  'ruby2.7',  'go1.x', 'java11') then 'ok'
```

And change in line 11 also:

```sql
    when runtime in ('nodejs14.x', 'nodejs12.x', 'nodejs10.x', 'python3.8', 'python3.7', 'python3.6', 'ruby2.5', 'ruby2.7', 'java11', 'java8', 'go1.x', 'dotnetcore2.1', 'dotnetcore3.1') then
```

to

```sql
    when runtime in ('nodejs16.x',  'python3.9',  'ruby2.7',  'go1.x', 'java11') then title || ' uses latest runtime - ' || runtime || '.'
```

So the whole `query/lambda/lambda_function_use_latest_runtime.sql` query now looks like:

```sql
select
  -- Required Columns
  arn as resource,
  case
    when package_type <> 'Zip' then 'skip'
    when runtime in ('nodejs16.x',  'python3.9',  'ruby2.7',  'go1.x', 'java11') then 'ok'
    else 'alarm'
  end as status,
  case
    when package_type <> 'Zip' then title || ' package type is ' || package_type || '.'
    when runtime in ('nodejs16.x',  'python3.9',  'ruby2.7',  'go1.x', 'java11') then title || ' uses latest runtime - ' || runtime || '.'
    else title || ' uses ' || runtime || ' which is not the latest version.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_lambda_function;

```

You can find the SQL file on [github](https://github.com/megaproaktiv/aws-community-projects/tree/main/steampipe-multi-account-query).

Steampipe noticed the changed query file and execute the querys again.

Wait for another execution in the terminal:

```bash
[ Wait    ] Dashboard execution started: aws_compliance.benchmark.foundational_security
[ Ready   ] Execution complete: aws_compliance.benchmark.foundational_security
```

In the updated browser windows you see all lambdas, multi-account.

Now the `compare.py` is ok:

![Compare Py](/img/2022/08/compare-py.png)

## More to play

For some IAM controls of the dashboard "AWS Foundational Security Best Practices" you need AWS IAM credential reports.

If you klick on "5 MFA should be enabled for all IAM users that have a console password" without the credential report, you get an error:

```log
...rpc error: code = Unknown desc = Credential report not available...
```

### Credential Reports

Call the report generation with the profile names:

```bash
aws iam generate-credential-report --profile megaproaktiv_dev
aws iam generate-credential-report --profile megaproaktiv_prod
aws iam generate-credential-report --profile megaproaktiv_test
```

Started:

```json
{
  "State": "STARTED",
  "Description": "No report exists. Starting a new report generation task"
}
```

Complete

```json
{
  "State": "COMPLETE"
}
```

Usually takes only seconds

## Even more to play

### Least Privileges

I started with `AWSPowerUserAccess` permission set. That is not least privileges. Creating least privileges IAM policies is hard, but there is a trick:

`iamlive` to the rescue.

With [iamlive](https://github.com/iann0036/iamlive) you can generate a IAM policy from real world AWS api calls:

In another cli windows, start:

```bash
iamlive --set-ini --mode proxy --sort-alphabetical --force-wildcard-resource --output-file policy.json
```

In the windows where you start the dashboard do:

```bash
export HTTP_PROXY=http://127.0.0.1:10080
export HTTPS_PROXY=http://127.0.0.1:10080
ca_bundle = ~/.iamlive/ca.pem
```

Then start the dashboard again.

This will take a few minutes, but you end up with a nice least privileges policy in the file `policy.json`.

You will find an example of the new policy on [github](https://github.com/megaproaktiv/aws-community-projects/tree/main/steampipe-multi-account-query).

## Conclusion

Fast ad-hoc Multi-Account queries of AWS resources can be done with steampipe. Once you get the configuration right, you can work with dashboards or html exports or junit export etc. For me this is an easy way to monitor AWS resources in addition to the AWS Configservice.

How to get no more Lambda end-of-support mails?
Use [go on aws](https://www.go-on-aws.com/), that is backward compatible and needs no runtime version updates.

For more AWS development stuff, follow me on twitter [@megaproaktiv](https://twitter.com/megaproaktiv)

## See also

- [Sample configuration files](https://github.com/megaproaktiv/aws-community-projects/tree/main/steampipe-multi-account-query)
- [Steampipe AWS Compliance Mod](https://hub.steampipe.io/mods/turbot/aws_compliance)

## Thanks to

Photo by <a href="https://unsplash.com/es/@alien_spaceship?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Swanson Chan</a> on <a href="https://unsplash.com/s/photos/octopus?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Unsplash</a>
