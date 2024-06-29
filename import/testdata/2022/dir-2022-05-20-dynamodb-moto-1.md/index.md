---
title: "Getting started with testing DynamoDB code in Python"
author: "Maurice Borgmeier"
date: 2022-05-20
toc: false
draft: false
image: "img/2022/05/glenn-carstens-peters-RLw-UC03Gwc-unsplash.jpg"
thumbnail: "img/2022/05/glenn-carstens-peters-RLw-UC03Gwc-unsplash.jpg"
categories: ["aws"]
tags: ["level-200", "dynamodb", "python", "moto", "testing"]
summary: Testing is one of the most critical activities in software development and using third-party APIs like DynamoDB in your code comes with challenges when writing tests. Today, I'll show you how you can start writing tests for code that accesses DynamoDB from Python.
---

Testing is one of the most critical activities in software development and, from my experience, also one of the first things thrown overboard when the deadline gets close. Using third-party APIs like DynamoDB in your code comes with challenges when writing tests. Today, I'll show you how you can start writing tests for code that accesses DynamoDB from Python.

We'll begin by installing the necessary dependencies to write our tests. To access DynamoDB, we'll use the AWS SDK for Python (boto3). The library [moto](https://github.com/spulec/moto) helps with mocking AWS services for tests, and [pytest](https://docs.pytest.org/) is a widespread module that allows writing tests in Python. We can install all of them like this.

```terminal
pip install boto3 moto pytest
```

Next, we'll begin by creating a demo project that we will test. I'm opting for a simple Lambda function that accepts an event, queries some data from DynamoDB, and stores an aggregate in DynamoDB. The general idea is to include a few ways of accessing DynamoDB to see how we can test them. This is what our project structure looks like (you can find all of the [code on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-moto)):

```text
.
├── dev-requirements.txt
├── requirements.txt
├── setup.py
├── src
│   └── lambda_handler.py
└── tests
    └── test_lambda_handler.py
```

Here's the Lambda handler from `src/lambda_handler.py`. It lists all transactions for a client, sums them up, and finally stores a summary item in DynamoDB.

```python
import os
import boto3
import boto3.dynamodb.conditions as conditions

ENV_TABLE_NAME = "TABLE_NAME"

def get_table_resource():

    dynamodb_resource = boto3.resource("dynamodb")
    table_name = os.environ[ENV_TABLE_NAME]
    return dynamodb_resource.Table(table_name)

def get_transactions_for_client(client_id: str) -> list:
    table = get_table_resource()

    # Get all items in the partition that start with TX#
    response = table.query(
        KeyConditionExpression=\
            conditions.Key("PK").eq(f"CLIENT#{client_id}") \
            & conditions.Key("SK").begins_with(f"TX#") 
    )
    
    return response["Items"]

def save_transaction_summary(summary_item: dict):

    # Add key information
    summary_item["PK"] = f"CLIENT#{summary_item['clientId']}"
    summary_item["SK"] = "SUMMARY"

    # store the item
    table = get_table_resource()
    table.put_item(Item=summary_item)

def lambda_handler(event, context):
    
    client_id = event["clientId"]
    
    client_transactions = get_transactions_for_client(client_id)

    total_sum = sum(tx["total"] for tx in client_transactions)

    summary_item = {
        "clientId": client_id,
        "totalSum": total_sum
    }
    save_transaction_summary(summary_item)

    return summary_item
```

There are now different approaches to testing this. This code tries to access DynamoDB to query and put data. We could write an integration test that uses a real DynamoDB table to test this, but that has a significant drawback. We're using a shared resource, which complicates this subject. Another approach is to *mock* DynamoDB (without insulting it), which is what we'll do today. A mock is a form of a test double. To our code, it looks like DynamoDB and behaves like DynamoDB, but it's not DynamoDB.

The Python package [moto](https://github.com/spulec/moto) bundles *mocks* for many AWS services, including DynamoDB. The DynamoDB mock behaves mainly like the service - we can create tables, add data, query data, remove data, and much more. Not all features are supported, though - the [documentation](https://docs.getmoto.org/en/latest/docs/services/dynamodb.html) lists the available API calls. We can use this in combination with [pytest](https://docs.pytest.org/) to create a simple test setup.

We begin by creating two *fixtures*. A [fixture](https://docs.pytest.org/en/7.1.x/explanation/fixtures.html#about-fixtures) provides context for a test. That means it can do something before and optionally after the test ends. In our case, the Lambda code needs the table name to be passed in from an environment variable, so we need to set this up - this is what `lambda_environment()` does. The code also assumes that the DynamoDB table already exists, so we need to create one. The `data_table()` fixture sets up a DynamoDB mock and then enlists the regular boto3 library to create a table. The `yield` keyword says that the setup is done, and the test can run now. After the `yield`, we could write code that runs after the test, but moto automatically tears down the data structures.

```python
@pytest.fixture
def lambda_environment():
    os.environ[lambda_handler.ENV_TABLE_NAME] = TABLE_NAME

@pytest.fixture
def data_table():
    with moto.mock_dynamodb():
        client = boto3.client("dynamodb")
        client.create_table(
            AttributeDefinitions=[
                {"AttributeName": "PK", "AttributeType": "S"},
                {"AttributeName": "SK", "AttributeType": "S"}
            ],
            TableName=TABLE_NAME,
            KeySchema=[
                {"AttributeName": "PK", "KeyType": "HASH"},
                {"AttributeName": "SK", "KeyType": "RANGE"}
            ],
            BillingMode="PAY_PER_REQUEST"
        )

        yield TABLE_NAME
```

The fixtures themselves don't test anything. They help us set up the actual tests. Below, you can see an example of a test. As parameters, we pass in the names of our fixtures, which tells pytest to initialize them. Then we invoke the lambda handler function with an event (`{"clientId": "ABC"}`). Since our table is empty, we expect the total sum to be 0 because we don't have any transactions for this client.

```python
def test_lambda_no_tx_client(lambda_environment, data_table):
    """Tests the lambda function for a client that has no transactions."""
    
    response = lambda_handler.lambda_handler({"clientId": "ABC"}, {})
    expected_sum = 0

    assert response["totalSum"] == expected_sum
    assert get_client_total_sum("ABC") == expected_sum
```

This works smoothly. Next, we want to make sure our code works when there are transactions in the table, to sum up. That means we need to add data to our table. No problem, we create a new fixture that builds upon the `data_table()`. You can see that the new fixture refers to the `data_table()` fixture and then puts items into the table.

```python
@pytest.fixture
def data_table_with_transactions(data_table):
    """Creates transactions for a client with a total of 9"""

    table = boto3.resource("dynamodb").Table(data_table)

    txs = [
        {"PK": "CLIENT#123", "SK": "TX#a", "total": 3},
        {"PK": "CLIENT#123", "SK": "TX#b", "total": 3},
        {"PK": "CLIENT#123", "SK": "TX#c", "total": 3},
    ]

    for tx in txs:
        table.put_item(Item=tx)
```

Next, we can write a test that asserts the sum of the transactions is correctly calculated as 9.

```python
def test_lambda_with_tx_client(lambda_environment, data_table_with_transactions):
    """
    Tests the lambda function for a client that has some transactions.
    Their total value is 9.
    """
    
    response = lambda_handler.lambda_handler({"clientId": "123"}, {})

    expected_sum = 9

    assert response["totalSum"] == expected_sum
    assert get_client_total_sum("123") == expected_sum
```

The `get_client_total_sum()` function is a helper function that queries the summary item in the table. Usually, this would be part of your data access layer, and you could reuse it here.

We can now run these tests from the console and measure the test coverage, which is pretty good.

![DynamoDB Moto Results](/img/2022/05/ddb_moto_test_results.png)

These tests are a good start that will allow you to begin refactoring the code in the Lambda function. They also provide coverage of the critical code paths. The unit it tests here is relatively large because it covers the whole module. You can also add additional tests for the individual functions, making it easier to focus on one particular function when you want to refactor it.

You can find the code I've shown here [on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-moto). Hopefully, this has been helpful to you and will allow you to start testing your own code. I'm looking forward to your questions and feedback.

&mdash; Maurice

---

**Further reading:**
- [The Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)

Cover Photo by [Glenn Carstens-Peters](https://unsplash.com/@glenncarstenspeters?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/checklist?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)
