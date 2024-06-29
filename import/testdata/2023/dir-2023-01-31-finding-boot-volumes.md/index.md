---
author: "Thomas Heinen"
title: "Finding EBS Boot Volumes"
date: 2023-01-31
image: "img/2023/01/pexels-tibe-de-kort-9951077-41.jpg"
thumbnail: "img/2023/01/pexels-tibe-de-kort-9951077-41.jpg"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "storage", "ebs", "level-500", "ruby", "aws-cli"]
---
Recently I got a question on how to find boot volumes of AWS instances. While I did not get any background on the purpose of this, I found the task interesting enough to dig a bit deeper. As in "down to the binary level". Interested?

<!--more-->
There are two different cases to check: are the volumes attached to an instance or are they some historic remains in an "available" state.

## Option 1: Attached

If volumes are attached, the task is easy. You just have to look at the volumes and if they are attached to an instance either as `/dev/xvda` or `/dev/sda1`. You do not even need a script for this, but can use an AWS CLI call with a [JMESPath](https://jmespath.org/)-based query:

```shell
aws ec2 describe-volumes --query 'Volumes[].Attachments[?Device == `/dev/xvda` || Device == `/dev/sda1`][VolumeId][][]'
```

This will list all volumes, filter for those which have `Attachments` to one of those two devices, only leave the volume ID and flatten this to a nicely parseable list:

```json
[
	"vol-29b1412a3d29c3ef0",
	"vol-02b4f228355b03bbf",
	"vol-55ee29db406f77b23",
	"vol-fd430a5139d351f93",
    "vol-f0da25b823de17d19"
]
```

There might be AMIs that have different root volumes, though. Before doing anything rash, check if there are instances that use a different convention.

## Option 2: Not Attached

Now, if you want to get the same information when EBS volumes are not attached - you seem to be out of luck: As every volume is just a virtual hard drive, there is no indicator in any AWS API on the type of volume. You can attach data volumes to instances and will only be able to see if it worked if you can connect to the instance or see error messages with the "instance screenshot" function.

But if you have regular snapshots (aka backups) configured for your instances, we can go deeper.

### EBS direct APIs
One little-known fact is, that AWS introduced the ability to read and write partial snapshots back in [December 2019](https://aws.amazon.com/about-aws/whats-new/2019/12/aws-launches-ebs-direct-apis-that-provide-read-access-to-ebs-snapshot-data-enabling-backup-providers-to-achieve-faster-backups-of-ebs-volumes-at-lower-costs/). This is intended for backup/recovery tools that can retrieve parts of a snapshot for file-level restores and similar actions. I already blogged about some possible [EBS direct API security implications](https://www.tecracer.com/blog/2021/09/be-aware-of-ebs-direct-apis.html) a while back.

For our use case, this allows us to access parts of the snapshots (you cannot access volumes in this way) to check for their contents. Why is this enough? It's called ...

## Partition Tables

If you ever had to debug boot issues or had some level of data corruption, you probably already touched partition tables. Simply put, these are reserved areas on a hard drive (virtual or physical) which give metadata about its contents. You find information about the size of its data areas, the used filesystems, the overall structure, and which partitions are enabled for booting an operating system. This is what we are searching for.

Two types of partition tables make up the majority of cases:

* [MBR](https://en.wikipedia.org/wiki/Master_boot_record): Created way back in 1983, with limited functionality, up to 2 Terabyte of disk
* [GPT](https://en.wikipedia.org/wiki/GUID_Partition_Table): Much more flexible, up to 75,600,000 Terabytes (should be enough for a while)
The key takeaway from this: You can check the partition table if there is any bootable operating system on the disk.

Before you dive into the actual binary specifications, I can tell you that in the frequently used "MBR" format (and on common AWS AMIs) this will be signaled by the hex value of `0x80` (128 in decimal) at byte `0x1BE` (446 in decimal) on the disk

### Putting it all together

So this is our workflow for non-attached volumes:

* iterate over all volumes
* find their most recent snapshot which includes the first block of the disk (let us call this "Block Zero")
* retrieve Block Zero with a single `ebs:GetSnapshotBlock` call
* if it is an MBR partition table
	* it is a boot/root volume if its byte `0x1BE` is `0x80`
* if it's a GPT table
	* check partition entries' [QWord attributes at `0x30` and bit 2](https://en.wikipedia.org/wiki/GUID_Partition_Table#Partition_entries_(LBA_2%E2%80%9333))

I did not encounter many GPT-based AMIs on Amazon yet, so let us stick to the easier-to-understand MBR variant.

## Ruby Example

From my blog entries, it should be obvious that I am most comfortable with Ruby.

As such, here is some quick code that will scan your current account for EBS volumes and return a JSON with information on if they are bootable, not bootable, or do not have snapshots (`unknown`).

```ruby
require 'aws-sdk-ebs'
require 'aws-sdk-ec2'
require 'json'

def ebs_volumes
  $ec2_client.describe_volumes.volumes
end

# Returns all snapshots, with the most current first
def snapshots_for_volume(volume_id)
  snapshots = $ec2_client.describe_snapshots(filters: [{ name: 'volume-id', values: [volume_id] }]).snapshots
  snapshots.sort_by(&:start_time).reverse
end

# Check snapshots for the latest one which includes the first block (index 0)
def latest_block_zero(snapshots)
  snapshots.each do |snapshot|
    snapshot_id = snapshot.snapshot_id
    blocks = $ebs_client.list_snapshot_blocks(snapshot_id: snapshot_id).blocks

    found = blocks.detect { |block| block.block_index.zero? }
    return { snapshot_id: snapshot_id, block: found } if found
  end
end

# Get contents of the first block only
def get_block(snapshot_id, block_token)
  {
    snapshot_id: snapshot_id,
    block: $ebs_client.get_snapshot_block(snapshot_id: snapshot_id, block_index: 0, block_token: block_token)
  }
end

def mbr?(block_zero)
  block_zero.block_data.rewind

  # MBR starts with this byte sequence
  return false if block_zero.block_data.getbyte != 0xEB
  return false if block_zero.block_data.getbyte != 0x63
  return false if block_zero.block_data.getbyte != 0x90

  true
end

def bootable?(block_zero)
  block_zero.block_data.rewind
  block_zero.block_data.pos = 0x1BE
  block_zero.block_data.getbyte == 0x80
end

###############

$ec2_client = Aws::EC2::Client.new
$ebs_client = Aws::EBS::Client.new
data = {}

ebs_volumes.each do |volume|
  volume_id = volume.volume_id
  snapshots = snapshots_for_volume(volume_id)

  if snapshots.empty?
    data[volume_id] = { type: 'unknown' }
    next
  end

  latest     = latest_block_zero(snapshots)
  block_zero = get_block(latest[:snapshot_id], latest[:block].block_token)

  # Ignoring the GPT complexities for brevity of code
  root = mbr?(block_zero[:block]) && bootable?(block_zero[:block])

  data[volume_id] = { type: root ? 'root' : 'data' }
end

print(data.to_json)
```

You can find this code, the Kaitai-based implementation, and various additional info in my [bootable-volumes repository on GitHub](https://github.com/tecracer-theinen/bootable-volumes)

## Summary

While we are going beyond the scope of AWS APIs, this shows neatly how you can automate tasks with the EBS direct APIs and a bit of binary voodoo. If you want to go deeper, you can use a file format generator like [Kaitai](https://kaitai.io) which has a registry of file formats and can compile parsers for nearly any programming language.

As the workflow for this is more involved (installing Java, the Kaitai Compiler, specific `.ksy` files, ...) I skipped over this in this blog post. This cleaner solution would also enable cleaner checks of all partitions on a disk - I did stick to the first one in this example.
