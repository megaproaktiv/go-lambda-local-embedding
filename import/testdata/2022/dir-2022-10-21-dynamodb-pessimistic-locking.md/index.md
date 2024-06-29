---
title: "Implementing Pessimistic Locking with DynamoDB and Python"
author: "Maurice Borgmeier"
date: 2022-10-21
toc: false
draft: false
image: "img/2022/10/crispin-jones-qeXdOn1TTpM-unsplash.jpg"
thumbnail: "img/2022/10/crispin-jones-qeXdOn1TTpM-unsplash.jpg"
categories: ["aws"]
tags: ["level-300", "dynamodb", "locking"]

---

I will show you how to implement pessimistic locking using Python with DynamoDB as our backend. Before we start, we'll review the basics and discuss some of the design criteria we're looking for.

In an earlier post, I outlined to you [how to implement optimistic locking using DynamoDB](https://aws-blog.com/2021/07/implementing-optimistic-locking-in-dynamodb-with-python.html). There, I explained some of the reasons why locking is useful and which issues it can prevent. If you're unfamiliar with the topic, I suggest you check that one out first.

Locking allows you to get exclusive access to a resource to access or modify it. There are many different kinds of locks with varying levels of exclusivity. Some locks make resources read-only, while others block access to them altogether. I mentioned optimistic locking above, and today, we'll focus on pessimistic locking. The main difference is how they respond to conflicts.

Optimistic locks work on the assumption that there are *usually* no conflicts when accessing a resource. They focus on *detecting* intermittent changes between reading a resource and modifying it. Their overhead is minimal; you need to redo your computation and try again whenever a collision is detected. Optimistic locking is a great fit if that's rare in your workload and recomputing is cheap.

Where optimistic locking focuses on collision detection, pessimistic locking employs a collision prevention approach. Before you try to modify a resource, you need to acquire a lock that gives you exclusive access to the resource until you release it. If the resource is already locked, you need to wait and retry at a later time until the lock is released. Collisions aren't possible in this approach, but you pay by adding the overhead of the locking infrastructure. That's often worth it if the cost of recomputing is high or collisions are likely.

In a nutshell, a pessimistic lock grants exclusive access to a resource and allows us not to worry about messy concurrent access. Let's talk a bit about how this works. The API is straightforward. It consists of two methods:

- `acquire_lock` that's used to try and get exclusive access. It can succeed or fail. If it fails, we must wait and retry until it's successful.
- `release_lock` that's used to tell the system we're done processing and allow others to use the resource again.

In the background, these functions need a data structure to track who has the lock and if there is a lock. We're going to use DynamoDB for that later. Additionally, we want to be able to support locks on multiple resources (while very useful, this introduces the problem of [deadlocks](https://en.wikipedia.org/wiki/Deadlock)). Also, the implementation will be used in the real world, where "*everything fails all the time*" (Werner Vogels), so we want to introduce some sort of timeout where a lock is released after a while if the locking entity no longer exists. Given these requirements, we arrive at the following method signatures for Python:

```python
def acquire_lock(
    resource_name: str, timeout_in_seconds: int, transaction_id: str
) -> bool:
    pass

def release_lock(resource_name: str, transaction_id: str) -> bool:
    pass
```

Let me explain. The `acquire_lock` function takes three arguments. The resource name is the name of the resource to be locked. Time in seconds is the number of seconds we wish to use the resource until our lock expires. The transaction id is a sequence of characters we can use twofold:

1. Identify in the table which transaction has the current lock.
2. Use as a kind-of password in the `release_lock` function to ensure only the lock's acquirer can release it.

The `release_lock` function requires two arguments. First, the resource's name to release the lock for, and second, the transaction ID used to lock it in the first place. This way, we ensure that it's not easy to release the lock if you didn't initially acquire it. It's meant to make it less likely to make mistakes, not to counteract malicious actors.

Both functions modify data in a DynamoDB table. In my implementation, I'm using a table with a partition and sort key, but a simple table with only a partition key would suffice. If these words are new to you, I suggest you check out [DynamoDB in 15 minutes](https://aws-blog.com/2021/03/dynamodb-in-15-minutes.html). Each lock will be an item in the table, and we use the `UpdateItem` API to modify it, which conveniently provides atomic operations on the item level. Let's now take a look at the implementation of `acquire_lock`:

```python
def acquire_lock(
    resource_name: str, timeout_in_seconds: int, transaction_id: str
) -> bool:

    dynamodb = boto3.resource("dynamodb")
    ex = dynamodb.meta.client.exceptions
    table = dynamodb.Table(TABLE_NAME)

    now = datetime.now().isoformat(timespec="seconds")
    new_timeout = (datetime.now() + timedelta(seconds=timeout_in_seconds)).isoformat(
        timespec="seconds"
    )

    try:

        table.update_item(
            Key={"PK": "LOCK", "SK": f"RES#{resource_name}"},
            UpdateExpression="SET #tx_id = :tx_id, #timeout = :timeout",
            ExpressionAttributeNames={
                "#tx_id": "transaction_id",
                "#timeout": "timeout",
            },
            ExpressionAttributeValues={
                ":tx_id": transaction_id,
                ":timeout": new_timeout,
            },
            ConditionExpression=conditions.Or(
                conditions.Attr("SK").not_exists(),  # New Item, i.e. no lock
                conditions.Attr("timeout").lt(now),  # Old lock is timed out
            ),
        )

        return True

    except ex.ConditionalCheckFailedException:
        # It's already locked
        return False
```

First, we compute the timestamps for the new timeout value and the current time, which will be used later. Then we perform a conditional `UpdateItem` operation that tries to store the lock item with the resource name as part of the primary key, the transaction id, and the timeout timestamp as attributes. This update item only succeeds if at least one of these conditions is fulfilled:

1. The item did not exist before. The update item API performs a create or update operation, and checking if part of the key didn't exist before tells us if it's a create and not an update.
2. The lock is expired. If the timeout value on the item is less than the current time, the lock is expired, and we're free to overwrite it and gain exclusive access to the resource.

If the update item call succeeds, we return True because the lock has been acquired. If it fails, an exception is raised, and we return False. In this case, the client will have to try obtaining the lock again later.

That's it for acquiring a lock. Let's discuss releasing it by looking at `release_lock`:

```python
def release_lock(resource_name: str, transaction_id: str) -> bool:

    dynamodb = boto3.resource("dynamodb")
    table = dynamodb.Table(TABLE_NAME)

    ex = dynamodb.meta.client.exceptions

    try:
        table.delete_item(
            Key={"PK": "LOCK", "SK": f"RES#{resource_name}"},
            ConditionExpression=conditions.Attr("transaction_id").eq(transaction_id),
        )
        return True

    except (ex.ConditionalCheckFailedException, ex.ResourceNotFoundException):
        return False
```

In this implementation, we're using a conditional delete to try and delete an existing lock. The delete operation succeeds if the transaction id stored in the item matches the transaction id from the function arguments. If that's not the case (ConditionalCheckFailedException) or the lock doesn't exist (ResourceNotFoundException), the operation fails.

To verify that all of this works as expected, I wrote several unit tests that use [moto to mock DynamoDB tables for testing](https://aws-blog.com/2022/05/getting-started-with-testing-dynamodb-code-in-python.html).

```python
def test_that_a_lock_can_be_acquired_if_none_exists():
    """
    Assert that a lock can be acquired if there is no pre-existing lock
    """
    # ...

def test_that_a_lock_can_be_acquired_if_the_old_is_expired():
    """
    Assert that a lock can be acquired if the pre-existing lock is expired.
    """
    # ...

def test_that_a_lock_is_rejected_if_another_exists():
    """
    Assert that no lock is granted if the resource is already locked.
    """
    # ...

def test_that_a_lock_can_be_release_with_the_tx_id():
    """
    Assert that a lock can be released if we know its transaction id.
    """
    # ...

def test_that_a_lock_cant_be_released_that_doesnt_exist():
    """
    Assert that a non-existent lock can't be released.
    """
	# ...

def test_that_a_lock_cant_be_released_if_the_tx_doesnt_match():
    """
    Assert that releasing a lock fails if we use an incorrect transaction id.
    """
    # ...
```

Let's talk about some limitations this implementation has. The Achilles' heel is clock synchronization. It uses the current system time to check if a lock is expired and to set the expiration time for new locks. If that time is out of sync between different clients, it may result in problems. Especially for short lock periods (less than a few seconds), the likelihood of failure increases. For the intended use case, I know that our lock durations will be in the minute range, and here the possibility of failure is low enough to ignore.

Another issue is that there is currently no way to extend the lock duration, although that wouldn't be too difficult to implement - it will be left as an exercise for the reader. The lock won't need to be extended for the intended use case.

Be sure to implement your clients in a way that they release the lock when they're done. Otherwise, the resource will be locked until the lock expires on its own. Clients should be fair.

You can find the [implementation, including the tests, on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-pessimistic-locking).

Hopefully, you learned something new today. For any questions, feedback, or concerns, feel free to reach out to me via the social media channels listed in my bio.

&mdash; Maurice

---

Photo by [Crispin Jones](https://unsplash.com/@cavespider?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/rusty-lock?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)