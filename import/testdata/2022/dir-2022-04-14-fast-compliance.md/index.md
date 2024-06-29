---
title: "About Optimizing for Speed: How to do complete AWS Security&Compliance Scans in 5 minutes"
author: "Gernot Glawe"
date: 2022-04-14
draft: false
image: "img/2022/04/coworker.jpeg"
thumbnail: "img/2022/04/coworker.jpeg"
toc: true
keywords:
  - security
  - compliance
  - go
  - python
  - rust
tags:
  - level-300
  - security
  - well-architected

categories: [aws]
---

The project steampipe uses a fast programing language and an intelligent caching approach outrunning prowler speed tenfold. While I tried to workaround prowlers limits I learned a lot about optimizing.

<!--more-->

## AWS security best practices assessments

To support security first, you should check your account against best practices and benchmarks. Two of the most used are the
[AWS Foundational Security Best Practices](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-fsbp-controls.html) from AWS and the CIS (Center for Internet Security) [AWS Benchmark](https://www.cisecurity.org/benchmark/amazon_web_services).

I will compare the widely used tool prowler (git 5.1k stars) with a newer one, steampipe (git 1.2k stars), which shows an auspicious approach.

## Prowler

Prowler introduces itself:

"Prowler is an Open Source security tool to perform AWS security best practices assessments, audits, incident response, continuous monitoring, hardening and forensics readiness. It contains more than 200 controls covering CIS, PCI-DSS, ISO27001, GDPR, HIPAA, FFIEC, SOC2, AWS FTR, ENS and custome security frameworks."

### Architecture

The application runs bash scripting with the AWS CLI as the base. Queries are done with a combination of bash pipes and the cli "query" syntax. This makes implementing controls easy at the beginning.

![Sequence of calls](/img/2022/04/diagram-1.svg.png)

The bash script waits after each AWS API request.

### Speed

The problem with that architecture is that a scan can take 30 minutes up to several hours.

Let's have a look at one check as an example:

### check122 as example

In `checks/check121` from the prowler repository, you will find this bash script:

#### Configuration section

The attributes of the check are stored in environment variables.

```bash
CHECK_ID_check122="1.22"
CHECK_TITLE_check122="[check122] Ensure IAM policies that allow full \"*:*\" administrative privileges are not created"
CHECK_SCORED_check122="SCORED"
CHECK_CIS_LEVEL_check122="LEVEL1"
CHECK_SEVERITY_check122="Medium"
CHECK_ASFF_TYPE_check122= "Software and Configuration Checks/Industry and Regulatory Standards/CIS AWS Foundations Benchmark"
CHECK_ASFF_RESOURCE_TYPE_check122="AwsIamPolicy"
CHECK_ALTERNATE_check122="check122"
CHECK_SERVICENAME_check122="iam"
CHECK_RISK_check122='IAM policies are the means by which privileges are granted to users; groups; or roles. It is recommended and considered a standard security advice to grant least privilege—that is; granting only the permissions required to perform a task. Determine what users need to do and then craft policies for them that let the users perform only those tasks instead of allowing full administrative privileges. Providing full administrative privileges instead of restricting to the minimum set of permissions that the user is required to do exposes the resources to potentially unwanted actions.'
CHECK_REMEDIATION_check122='It is more secure to start with a minimum set of permissions and grant additional permissions as necessary; rather than starting with permissions that are too lenient and then trying to tighten them later. List policies an analyze if permissions are the least possible to conduct business activities.'
CHECK_DOC_check122='http://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html'
CHECK_CAF_EPIC_check122='IAM'
```

#### Code section

```bash
  1 check122(){
  2   # "Ensure IAM policies that allow full \"*:*\" administrative privileges are not created (Scored)"
  3   LIST_CUSTOM_POLICIES=$($AWSCLI iam list-policies --output text $PROFILE_OPT --region $REGION --scope Local --query 'Policies[*].[Arn,Defau    ltVersionId]' | grep -v -e '^None$' | awk -F '\t' '{print $1","$2"\n"}')
  4   if [[ $LIST_CUSTOM_POLICIES ]]; then
  5     for policy in $LIST_CUSTOM_POLICIES; do
  6       POLICY_ARN=$(echo $policy | awk -F ',' '{print $1}')
  7       POLICY_VERSION=$(echo $policy | awk -F ',' '{print $2}')
  8       POLICY_WITH_FULL=$($AWSCLI iam get-policy-version --output text --policy-arn $POLICY_ARN --version-id $POLICY_VERSION --query "[Policy    Version.Document.Statement] | [] | [?Action!=null] | [?Effect == 'Allow' && Resource == '*' && Action == '*']" $PROFILE_OPT --region $REGION    )
  9       if [[ $POLICY_WITH_FULL ]]; then
 10         POLICIES_ALLOW_LIST="$POLICIES_ALLOW_LIST $POLICY_ARN"
 11       fi
 12     done
 13     if [[ $POLICIES_ALLOW_LIST ]]; then
 14       for policy in $POLICIES_ALLOW_LIST; do
 15         textFail "$REGION: Policy $policy allows \"*:*\"" "$REGION" "$policy"
 16       done
 17     else
 18         textPass "$REGION: No custom policy found that allow full \"*:*\" administrative privileges" "$REGION"
 19     fi
 20   else
 21     textPass "$REGION: No custom policies found" "$REGION"
 22   fi
 23 }
```

The code does the following:

1\. List policies (line 3)

```bash
$AWSCLI iam list-policies
```

2\. Loop (line 5-12)

```bash
$AWSCLI iam get-policy-version
```

### AWS cli speed

With python for _each_ call, the python interpreter starts, taking time on each call.

Let me show you what I mean:

```bash
time aws iam list-policies
aws iam list-policies  0,93s user 0,13s system 15% cpu 6,894 total
```

It takes about 7 seconds. Now "list-policies" is a timewise costly operation, so I look at a simpler one, S3 list.

To see which part is python time and which AWS time, I compare "aws s3 ls" with a faster go application.

This is the AWS python cli:

```bash
time aws s3 ls
aws s3 ls  0,46s user 0,07s system 72% cpu 0,726 total
```

Now you could say, the terminal output takes some time. We eliminate that:

```bash
time aws s3 ls >/dev/null
aws s3 ls > /dev/null  0,43s user 0,07s system 79% cpu 0,635 total
```

So **0.635 seconds in total** for AWS CLI/Python

![Bash Python cli](/img/2022/04/cli.svg.png)

In Contrast, a small GO program takes less time:

```bash
time ./s3list >/dev/null
./s3list > /dev/null  0,01s user 0,01s system 11% cpu 0,183 total
```

Where the main code is like:

```go
res, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
for _, bucket := range res.Buckets {
        bucketarray = append(bucketarray, bucket.Name)
}
return bucketarray, nil
```

So **0.183 seconds in total** for a compiled GO program.

Some of you may say: Rust is the new speed king - ok, let's try:

Rust _could be_ faster.
But, in the moment, the rust SDK is still in beta, so it takes more time than GO, but is faster than python.

```bash
time target/debug/rust-s3 >/dev/null
target/debug/rust-s3 > /dev/null  0,15s user 0,03s system 43% cpu 0,403 total
```

So **0.403 seconds in total** for a compiled Rust program, created with `cargo build`.

This is the rust code:

```rust
use aws_config::meta::region::RegionProviderChain;
use aws_sdk_s3::{Client, Error};
#[tokio::main]
async fn main() -> Result<(), Error> {
    let region_provider = RegionProviderChain::default_provider().or_else("eu-central-1");
    let config = aws_config::from_env().region(region_provider).load().await;
    let client = Client::new(&config);
    if let Err(e) = show_buckets(&client).await
    {
        println!("{:?}", e);
    };
    Ok(())
}
async fn show_buckets( client: &Client) -> Result<(), Error> {
    let resp = client.list_buckets().send().await?;
    let buckets = resp.buckets().unwrap_or_default();
    for bucket in buckets {
        println!("{}", bucket.name().unwrap_or_default());
    }
    println!();
    Ok(())
}
```

## Optimization

As you see, a compiled GO program is much faster than the python script. So I tried to optimize the cli with a GO cache.

The project is [awclip](https://github.com/megaproaktiv/awclip). Please just see it as an experiment; it worked somehow but did not create the desired results.

What I found was:

- AWS CLI is doing _many more_ things than just calling the service
- AWS CLI is not always json.

Looking at line 8 of the check:

```bash
--query "[Policy    Version.Document.Statement]
```

This works on the json, but important APIS like [EC2 API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Welcome.html), [S3 API](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListBuckets.html) and [IAM](https://docs.aws.amazon.com/IAM/latest/APIReference/welcome.html) reponse with _XML_, not json.

So replacing the "query" function with an own program is not easy.

The most promising approach was to prefetch regions in parallel. Prowler is doing everything sequentially.

But with about a day's work, I only got 10% better speed out of prowler. That made me realize that my thinking was wrong.

I was optimizing _locally_, not _globally_.

If you read the book "The Goal", it says:

"We are not concerned with local optimums."

So my approach to only replace AWS cli as a local optimum lead nowhere - back to the drawing board. After some research, I discovered a project which used a global optimizing approach: Steampipe

## Steampipe - a different approach

Steampipe uses a Postgres Foreign Data Wrapper to present data and services from external systems as database tables.

![Steampipe](/img/2022/04/steampipe-arch.png)

The queries are not run on a JSON like in AWS cli. The AWS data is presented in tables, and the queries work directly on those tables. All results are cached and the queries run on the local postgres table, not as AWS service calls.

So when I look at a similar control like `check-122`, here called "Control: 1 IAM policies should not allow full '\*' administrative privileges", the query is written in SQL.

The configuration (`foundational_security/iam.sp`) is separated from the logic in a separate file and looks like this:

```bash
control "foundational_security_iam_1" {
  title         = "1 IAM policies should not allow full '*' administrative privileges"
  description   = "This control checks whether the default version of IAM policies (also known as customer managed policies) has administrator access that includes a statement with 'Effect': 'Allow' with 'Action': '*' over 'Resource': '*'. The control only checks the customer managed policies that you create. It does not check inline and AWS managed policies."
  severity      = "high"
  sql           = query.iam_custom_policy_no_star_star.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_1.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_1"
    foundational_security_category = "secure_access_management"
  })
}
```

This query is located in `query/iam/iam_policy_no_star_star.sql` in the [github repository](https://github.com/turbot/steampipe-mod-aws-compliance.git)

With prowler it is a query with AWS cli and pipe some Bash logic:

```bash
--query "[PolicyVersion.Document.Statement] | [] | [?Action!=null] | [?Effect == 'Allow' && Resource == '*' && Action == '*']"
```

With steampipe the syntax is sql:

```sql
with bad_policies as (
  select
    arn,
    count(*) as num_bad_statements
  from
    aws_iam_policy,
    jsonb_array_elements(policy_std -> 'Statement') as s,
    jsonb_array_elements_text(s -> 'Resource') as resource,
    jsonb_array_elements_text(s -> 'Action') as action
  where
    s ->> 'Effect' = 'Allow'
    and resource = '*'
    and (
      (action = '*'
      or action = '*:*'
      )
  )
  group by
    arn
)
```

The bash text comparism `Effect == 'Allow'` becomes `where s ->> 'Effect' = 'Allow'`.

### Developing SQL queries

You can develop the sql statements with the `steampipe query` command, which gives you a local sql query tool
The first step is that with the `.tables` command you get all AWS tables after installing the AWS plugin.

![tables](/img/2022/04/tables.png)

### inspect aws_iam_policy

The second step is to see which fields you may query:

![iam_policy](/img/2022/04/iam_policy.png)

The `.inspect` command shows the fields.

You may start to experiment with the queries:

```bash
> select policy_std from aws_iam_policy
+----------------------------------------------------------------------------------------------------------------------
| policy_std
+----------------------------------------------------------------------------------------------------------------------
| {"Statement":[{"Action":["logs:createloggroup",...
```

The functions to work with JSON make it possible to interpret the policy documents, see:
`jsonb_array_elements_text(s -> 'Action') as action`
in the following statement:

```sql
jsonb_array_elements(policy_std -> 'Statement') as s,
jsonb_array_elements_text(s -> 'Resource') as resource,
jsonb_array_elements_text(s -> 'Action') as action
where
  s ->> 'Effect' = 'Allow'
  and resource = '*'
  and (
    (action = '*'
    or action = '*:*'
    )
)
```

Now we know how a control is built, let's look at the speed.

## Speed for the single control

### Speed with prowler

Running the single control with prowler takes more than 2 minutes:

![single-prowler](/img/2022/04/single-prowler.png)

### Speed with steampipe

With steampipe it is down to 23 seconds while giving the same results.

![single-steampipe](/img/2022/04/single-steampipe.png)

## Speed for whole scan

When doing a complete scan, prowler takes about **30** minutes for one region, steampipe is running in **3** minutes.

_Maybe_ you want to try this yourself - well _CloudShell_ is your friend.

Here is a complete small walkthrough to check your account!

## Walktrough getting started

The overall steps are:

- Download and install Steampipe
- Install the AWS plugin
- Clone repo steampipe-mod-aws-compliance
- Generate your AWS credential report
- Configure regions (optional)
- Run all benchmarks
- Download report file

In detail:

1\. Login into your AWS account with admin rights

2\. Open a cloudshell

![cloudshell](/img/2022/04/cloudshell.png)

3\. Execute these commands:

```bash
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
steampipe plugin install steampipe
steampipe plugin install aws
git clone https://github.com/turbot/steampipe-mod-aws-compliance.git
cd steampipe-mod-aws-compliance
aws iam generate-credential-report
```

4\. (Optional) Edit region config

If you just want to scan specific region(s), you can set this in the configuration\_

```bash
vi ~/.steampipe/config/aws.spc
```

```bash
  1 connection "aws" {
  2   plugin    = "aws"
  3
  4   regions     = ["eu-central-1","us-east-1"]
  5
  6 }
```

You know that you do not copy the line numbers, do you ;)

5\. Wait for credential report

```bash
aws iam generate-credential-report
```

You get this when the report is running:

```json
{
  "State": "STARTED",
  "Description": "No report exists. Starting a new report generation task"
}
```

And this, when its done:

```json
{
  "State": "COMPLETE"
}
```

6\. Execute complete check with html output

```bash
steampipe check all --export=html
```

This should take only 3-5 minutes.

You see the process running:

![running](/img/2022/04/running.png)

7\. Download report

Look for all-$date-$time.html. like `all-20220412-150100.html`

In the Actions menu of cloudshell you can "Download" the html file:

![download](/img/2022/04/download.png)

You have to write the whole path, like this:

![path](/img/2022/04/path.png)

Now you have the full report:

![report](/img/2022/04/report.png)

## Conclusion

### Security Scan

There are many good open-source AWS compliance and security scanning tools on github.

With a full report - including installation on cloud shell - in about 5 minutes, SteamPipe is the new shooting star in my tools portfolio. We will see in the long run.

You should really take the 10 minutes to try it!

### Optimizing

What I have learned:

- Prefer global before local optimizing.
- Tools and languages have restrictions.
- Look for alternatives instead of spending much time on workarounds.
- Compiled is faster than interpreted.
- Do not trust general statements. Write the smallest proof of concept possible to challenge your belief.

## Feedback & discussion

For discussion please contact me on twitter @megaproaktiv

## Learn more GO

Want to know more about using GOLANG on AWS? - Learn GO on AWS: [here](https://www.go-on-aws.com/)

## Sources

- Goldratt, Eliyahu M.; Jeff Cox. The Goal: A Process of Ongoing Improvement . North River Press. Kindle-Version.
- [Steampipe check](https://steampipe.io/docs/reference/cli/check)
- [Steampipe Hub](https://hub.steampipe.io/)
- [Prowler](https://github.com/prowler-cloud/prowler)
