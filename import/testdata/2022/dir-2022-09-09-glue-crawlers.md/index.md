---
title: "Glue Crawlers: No GetObject, No Problem"
author: "André Reinecke, Maurice Borgmeier"
date: 2022-09-09
toc: false
draft: false
image: "img/2022/09/nitish-meena-RWAIyGmgHTQ-unsplash.jpg"
thumbnail: "img/2022/09/nitish-meena-RWAIyGmgHTQ-unsplash.jpg"
categories: ["aws"]
tags: ["level-300", "glue", "crawler"]

---

This is the story of how we accidentally learned more about the internals of Glue Crawlers than we ever wanted to know. Once upon a time (a few days ago), André and I were debugging a crawler that didn't do what it was supposed to. Before we dive into that, maybe some background on Crawlers first.

Glue Crawlers are used to create tables in the Glue Data Catalog. They crawl, i.e., analyze one or more data sources like S3 buckets, make educated guesses about the structure of files and directories they find, and, based on that, create tables in the data catalog. In this process, they also read the files' content to figure out how they're structured. In our experience, Crawlers can be very useful but frustrating because it's not transparent how it figures out the table structure. When it goes wrong, it's often down to guesswork and experience. You can learn more about detection of schemas [here](https://aws.amazon.com/premiumsupport/knowledge-center/glue-crawler-detect-schema/?nc1=h_ls).

With this background, let's set the stage. Like many IT stories, this one plays out in an enterprise environment (where simple things get ~~complicated~~ secure). A glue crawler was supposed to create tables from a well-structured S3 bucket where the table name is at a fixed level in the object key. You can imagine something like this:

```
s3://bucket-name/system-name/<table-name>/partition=value/file.parquet
```

Since the table name is guaranteed to be unique in this environment, the Crawler was supposed to start in the bucket and take the table name level as the table name. Unfortunately, that wasn't working, and the Crawler didn't write any logs due to permission issues. At that point, the Crawler had created one table.

During debugging, the policy was replaced with the `AWSGlueServiceRole` [policy](https://docs.aws.amazon.com/glue/latest/dg/create-service-policy.html), which contains (among others) these S3 permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetBucketLocation",
                "s3:ListBucket",
                "s3:ListAllMyBuckets",
                "s3:GetBucketAcl",
                // ...
            ],
            "Resource": [
                "*"
            ]
        },
        // ...
    ]
}
```

The `s3:ListBucket` permission allows Glue to list all objects in all buckets this account has access to (yes, `s3:ListBucket` maps to the list object API, [it's intuitive](https://aws.amazon.com/premiumsupport/knowledge-center/s3-access-denied-listobjects-sync/#:~:text=Note%3A%20s3%3AListBucket%20is%20the%20name%20of%20the%20permission%20that%20allows%20a%20user%20to%20list%20the%20objects%20in%20a%20bucket.%20ListObjectsV2%20is%20the%20name%20of%20the%20API%20call%20that%20lists%20the%20objects%20in%20a%20bucket.)). What this policy doesn't contain, is `s3:GetObject` permissions for all buckets (which is good). This is where the weird behavior started.

During debugging, the table that the Crawler had created earlier was deleted to ensure it was still doing things. **The Crawler even created that specific table again** - with columns, datatypes, and everything - but no other tables (for newly put or edited files). At first, this didn't seem weird because nobody noticed that the policy's KMS and `s3:GetObject` permissions were missing.

Once the missing-log problem had been fixed, we could see S3 Access Denied messages in the logs. From there, we quickly figured out that `kms:Decrypt` and `s3:GetObject` were missing in the policy and added those. Afterward, **all** the tables got created as expected.

Then we were thinking to ourselves: wait a minute, how could the Crawler create the other table earlier without being able to read the object and figure out the structure? Is it somehow evaluating metadata on the object or bypassing the permissions? Later it turned out that neither of those was true. To investigate further, we set up a similar setup in another account to escape the enterprise lockdown.

1. We created an S3-Bucket with a well-structured format and some CSV and Parquet files
2. We created a Glue Crawler (`$glueCrawler`) with full access permissions that was then able to discover all the tables.
3. Next, we changed the policy on `$glueCrawler` to the limited policy from the `AWSGlueServiceRole` mentioned above. The policy that doesn't have `s3:GetObject` for our created S3 bucket.
4. Then we deleted all the tables from the Data Catalog that the Crawler had discovered earlier and re-run `$glueCrawler`.
5. The Crawler could create all the tables like before even though it couldn't read the objects.

That confirmed the behavior we had seen earlier in the enterprise environment. We suspected some caching was going on because the Crawler could still list the objects in the bucket (`s3:ListBucket`). To confirm we were getting cached results, we continued the experiment:

6. We added another CSV to the bucket in a new well-defined path
7. We deleted the tables from the Data Catalog and re-ran `$glueCrawler`.
8. The Crawler re-created all the tables from before, **but not the new one** - the logs showed access denied messages.

After that, we were pretty sure that we were observing caching behavior - the Crawler was giving us tables it had seen before but could not add to the cache. Then we wanted to find out if it makes mistakes:

9. We deleted the data for one of the original tables and then deleted all the tables from the Glue Catalog
10. We re-ran `$glueCrawler` and observed that the new table wasn't created, but all the old ones that still existed were created - the deleted one didn't get re-created.

From that, we learned that the caching seems quite effective and, in this scenario, at least correct. Next, we were curious if the cache is local to the Crawler or shared between Crawlers:

11. We created a new Crawler, `$anotherCrawler` with the same limited role and configuration
12. We deleted the tables from the data catalog
13. We ran `$anotherCrawler`, and it created **no** tables.

That lets us conclude that the cache is indeed local to the Crawler and not shared between multiple Crawlers.

We learned a few things about Crawlers through this debugging session:

- Crawlers have external state that's persisted outside of a run but scoped to a single crawler
- Crawlers seem to use the ListObjects API to list the available objects in the prefix and recreate a table if it already knows the file without reading it again
- That way, it may partially create tables even though it no longer has access to the underlying data

At first, we were unsure what to think about this implementation and then decided that it's pretty neat (even though it had cost us a few hours):

- This caching mechanism avoids re-processing unchanged data (and the costs associated with that)
- The Crawler doesn't create wrong results (e.g., deleted tables)
- The Crawler is relatively fault-tolerant if it loses access to the underlying data

After this experience, we're left with mixed feelings. A mistake during debugging let us learn about this behavior, which is kind of neat. On the other hand, it also feels like *unexpected magic* that the Crawler can create a table without being able to read the data anymore.

Hopefully, you enjoyed learning with us!

André & Maurice

---

Title Photo by [Nitish Meena](https://unsplash.com/@nitishm?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/discover?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)
