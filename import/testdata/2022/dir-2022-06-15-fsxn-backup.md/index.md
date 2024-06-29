---
author: "Thomas Heinen"
title: "FSx for ONTAP Backups"
date: 2022-06-15
image: "img/2022/06/xavi-cabrera-kn-UmDZQDjM-unsplash.png"
thumbnail: "img/2022/06/xavi-cabrera-kn-UmDZQDjM-unsplash.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "fsx", "backup", "netapp"]
---

In our [FSx for NetApp ONTAP series](https://aws-blog.de/tags/fsx.html), we continue to one of the most vital topics: Backups. But did you know there are two types of backup with this service? Let's compare the native backup and AWS Backup in this post.

<!--more-->

## General

As you might remember, AWS backup works differently from other backup solutions because we are talking about a snapshot-based approach and not a file-based one.

This approach has several advantages: no need for a backup agent, instant backup windows, and no additional system load - as backup is done on the storage and block level.

Consequently, your first snapshot contains all data, and all subsequent ones only have the differences from the one before. If you did classical images, you would need to pay a multiple of your hard drive size. But with the techniques we are using on AWS, you only pay for the data you change.

With this knowledge, we arrive at a formula for how you calculate your aggregate backup space:

`raw_capacity + raw_capacity * daily_change_rate/100 * retention_time`

__Quick example:__ If you have 100 TB of capacity and a daily change rate of 3% with 90 days of retention, you can expect 370 TB of backup size (and cost).

## FSx for NetApp ONTAP Native Backup

The first type of backup is the Amazon FSx for NetApp ONTAP (FSxN) native one: it simply uses the described snapshot technology to keep a local list of previous snapshots. It runs daily in a defined period and has a default retention of seven days (1 - 90 days are possible).

A significant caveat is that these automatic backups are tied to their volume. In NetApp terminology, these are called FlexVols, and snapshots use the inherent capabilities of the [WAFL filesystem](https://en.wikipedia.org/wiki/Write_Anywhere_File_Layout) not to copy data but to create pointers to the used blocks on disk. As soon as a snapshot is requested, there will be another list of pointers to the data blocks at this point. As soon as the original volume changes, there will be new blocks, and the maps between the actual volume and the snapshot will diverge.

![FSxN Snapshots](/img/2022/06/fsxn-backup-snapshots.png)

But that also means another thing: if you delete the FlexVol, __all those automatic snapshots will be deleted as well__.

There is a big difference with manual, user-initiated backups: These will be completely standalone, and volume deletion will not affect them.

## AWS Backup

While [AWS Backup](https://aws.amazon.com/backup/) uses the same mechanics as the native backup, its backups are _always_ independent of the originating volume. It also includes more scheduling options:

- adjustable frequency between hourly and monthly backups
- layering different policies allows for a classic [Grandfather-Father-Son (GFS) backup](https://backup.ninja/news/grandfatherfatherson-gfs-backup-strategy), not only rolling snapshots
- the ability to copy into different regions for disaster recovery scenarios
- WORM ability to lock contents against deletion (which might be a good idea with ransomware threats rising)

## Summary

After learning about the differences, you should be able to determine which type of backup is more appropriate for your workload. Most people will generally prefer the versatility and robustness of AWS Backup to the native technologies.

