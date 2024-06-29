---
title: "Teaching boto3 to store floats and datetime objects in DynamoDB"
author: "Maurice Borgmeier"
date: 2023-09-07
toc: false
draft: false
image: "img/2023/09/jamie-matocinos-rW00Wu_CeYA-unsplash.jpg"
thumbnail: "img/2023/09/jamie-matocinos-rW00Wu_CeYA-unsplash.jpg"
categories: ["aws"]
tags: ["level-400", "dynamodb", "boto3", "python"]
summary: |
  In this blog post, we'll explore how you can teach the DynamoDB Table resource in boto3 (and the client) to store and retrieve Python's datetime and float objects, which they can't do natively. We'll also discuss why you should or shouldn't do that.
---

The resource API in boto3 is a higher-level abstraction that lets you interact with AWS services in a way that feels more pythonic. This means you can use language idioms that make the code easier to understand, and boto3 translates that into the underlying representation that the AWS API likes. Recently, Lee Hannigan wrote about "[Exploring Amazon DynamoDB SDK clients](https://aws.amazon.com/blogs/database/exploring-amazon-dynamodb-sdk-clients/)" on the AWS blog, which is a good introduction if you're unfamiliar.

One of the quirks of this API when working with DynamoDB is that it only natively works with some data types and doesn't support others, like commonly used `float` or `datetime` objects.

```python
import boto3

table = boto3.resource("dynamodb").Table("data")
table.put_item(
    Item={"PK": "pk", "SK": "sk", "float": 3.14159265359}
)
```

Running this results in:

> TypeError: Float types are not supported. Use Decimal types instead.

Trying the same with a datetime object will give you the following:

> TypeError: Unsupported type "<class 'datetime.datetime'>" for value ...

This can be quite annoying because it makes you wonder why the high-level API isn't able to deal with these common data types. Part of the reason for this is most likely that floats in Python can be counter-intuitive, so `Decimal` is a better data type if you want numbers to behave as non-computer-scientists expect it. To learn more about these complexities, check out [this discussion on GitHub about implementing float support in boto3](https://github.com/boto/boto3/pull/2699) and the [Python documentation on the subject](https://docs.python.org/3/tutorial/floatingpoint.html). Additionally, DynamoDB has no native `DateTime` data type, so there is no straightforward mapping.

In order to fix these shortcomings, we can extend boto3 to teach it how to deal with more data types than it can on its own. The code, including some tests for it is [available on GitHub](https://github.com/MauriceBrg/aws-blog.de-projects/blob/master/dynamodb-more-types/test_type_serializer.py), and you can jump straight into that if you're impatient, but we're also going to walk through it and talk about some background info that may interest you.

DynamoDB uses an extended JSON syntax under the hood that explicitly encodes data types and also supports additional collections, such as sets that native JSON doesn't support. It looks like this:

```json
{
  "PK": {
    "S": "item_with"
  },
  "SK": {
    "S": "float_data"
  },
  "binary_data": {
    "B": "SGVsbG8gV29ybGQh"
  }
}
```

If you use the low-level client API, this is what you'll work with, and it's kind of annoying. The resource API hides this implementation detail (somewhat) by using the `TypeSerializer` and `TypeDeserializer` from the [`boto3.dynamodb.types`](https://boto3.amazonaws.com/v1/documentation/api/latest/_modules/boto3/dynamodb/types.html) module. Our problem is that this serializer and deserializer have no clue how to translate `float` and `datetime` objects.

Before we extend these, let's talk about storing our floats and datetimes in DynamoDB. For Floats, the number (`N`) data type looks promising, but if we use that, the deserializer would automatically decode that as `Decimal`, which we don't want and we don't want to break the existing implementation that can handle decimals just fine. On the other hand, the binary type (`B`) allows us to encode any information we like. That's why I chose to encode a `float` as something like `binary("FL:<number>")`

This means we convert the float to a string, add the `FL:` prefix, and convert this string to `bytes` in Python, which we then store in the binary data type in DynamoDB. Why binary and not string? Good question. I'll get back to that in a second. After teaching the serializer how to encode the float as binary information, we also need to extend the deserializer to translate the value back to a float.

I didn't want to mess with the string deserializer as that's commonly used. In my experience, the binary data type is rarely used, so making the deserialization a bit more complex shouldn't affect everything else too much. The idea is that the decoder checks if the first three bytes decoded are `FL:`, and if that's the case, we take the rest and convert it to a float. If we did that with a string field, we'd have to make sure that none of our other strings start with `FL:` - for binary information, this is technically also true but _much_ less likely.

For datetime objects, I chose to go with a similar solution. Since there is no native data type for datetime information, I decided to store the information as an [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) encoded string with the `DT:` prefix in a binary object.

With this background information, let's check out the new `CustomSerializer`:

```python
import typing
from datetime import datetime
from boto3.dynamodb.types import TypeSerializer

class CustomSerializer(TypeSerializer):
    """
    Thin wrapper around the original TypeSerializer that teaches it to:
    - Serialize datetime objects as ISO8601 strings and stores them as binary info.
    - Serialize float objects as strings and stores them as binary info.
    - Deal with the above as part of sets
    """

    def _serialize_datetime(self, value) -> typing.Dict[str, bytes]:
        return {"B": f"DT:{value.isoformat(timespec='microseconds')}".encode("utf-8")}

    def _serialize_float(self, value) -> typing.Dict[str, bytes]:
        return {"B": f"FL:{str(value)}".encode("utf-8")}

    def serialize(self, value) -> typing.Dict[str, typing.Any]:
        try:
            return super().serialize(value)
        except TypeError as err:

            if isinstance(value, datetime):
                return self._serialize_datetime(value)

            if isinstance(value, float):
                return self._serialize_float(value)

            if isinstance(value, Set):
                return {
                    "BS": [
                        self.serialize(v)["B"]  # Extract the bytes
                        for v in value
                    ]
                }

            # A type that the reference implementation and we
            # can't handle
            raise err
```

As you can see, the implementation inherits from the original `TypeSerializer` because I didn't feel like re-implementing stuff that works perfectly fine. I overrode the `serialize` method and decided that it was easiest to first call the original implementation and then deal with `TypeError`s should they arise. There are only three things we need to check for here:

1. Is the `value` a `datetime` object? If yes, serialize it as discussed above.
2. Is the `value` a `float` object? If yes, serialize it as discussed above.
3. Is the `value` a set? If yes, call the serializer on all set elements and extract the encoded `bytes`.

Now that our `CustomSerializer` can translate floats and datetimes into something DynamoDB can deal with, we need a `CustomDeserializer` to translate our binary information back to `float` and `datetime` objects.

```python
import typing
from datetime import datetime
from boto3.dynamodb.types import TypeDeserializer

class CustomDeserializer(TypeDeserializer):
    """
    Thin wrapper around the original TypeDeserializer that teaches it to:
    - Deserialize datetime objects from specially encoded binary data.
    - Deserialize float objects from specially encoded binary data.
    """

    def _deserialize_b(self, value: bytes):
        """
        Overwrites the private method to deserialize binary information.
        """
        if value[:3].decode("utf-8") == "DT:":
            return datetime.fromisoformat(value.decode("utf-8").removeprefix("DT:"))
        if value[:3].decode("utf-8") == "FL:":
            return float(value.decode("utf-8").removeprefix("FL:"))

        return super()._deserialize_b(value)
```

The custom deserializer is a bit dirty because we overwrite the original implementation's private method `_deserialize_b`. This means it relies on an implementation detail that may (although it's not very likely) change. The method is called when the deserializer tries to deserialize a binary data type. We just check if the first three bytes interpreted as a `utf-8` string equal one of our known prefixes, and if that's the case, deserialize the bytes to `float` or `datetime`. If our prefix isn't there, we hand it off to the original implementation.

So, how do we get boto3 to use our fancy new `CustomSerializer` and `CustomDeserializer` instead of the built-in implementations? I struggled with that a little bit and had to dig through the botocore/boto3 implementation to understand how it works. Internally, the (De)Serializer is wired into the [botocore event system](https://botocore.amazonaws.com/v1/documentation/api/latest/topics/events.html) which is frankly poorly documented. In a nutshell, it allows you to hook into the flow of data through the SDK at specific points in the lifecycle of an API call and potentially modify the data.

Specifically, there's a `TransformationInjector` in the `boto3.dynamodb.transform` module that takes care of invoking the original (De)Serializer at the right point in time. In order to make it call our own (De)Serializer, I created a subclass of that one and instantiated our (de)serializers.

```python
from boto3.dynamodb.transform import TransformationInjector

class CustomTransformationInjector(TransformationInjector):
    """
    Thin wrapper around the Transformation Injector that uses
    our serializer/deserializer.
    """

    def __init__(
        self,
        transformer=None,
        condition_builder=None,
        serializer=None,
        deserializer=None,
    ):
        super().__init__(
            transformer, condition_builder, CustomSerializer(), CustomDeserializer()
        )
```

It really doesn't do more than statically using our (de)serializers instead of the built-in ones. I didn't want to deal with or mess with all the data structures involved here, so the rest stayed vanilla.

Next, we need to kick out the old (de)serializers and use our own. This means we need to unregister the old handlers and register our new ones. The best place to do that is the `boto3.Session` object that can be used to create clients and resources. You could also do it for individual clients or resources, but then it won't be inherited from any resources they create. It's also important to update the registration before instantiating clients and resources, as it seems like all the handlers are copied at instantiation time.

```python
def build_boto_session() -> boto3.Session:
    """
    Build a session object that replaces the DynamoDB serializers.

    NOTE: It's important that the registration/unregistration
          happens before any resource or client is instantiated
          from the session as those are copied based on the
          state of the session at the time of instantiating
          the client/resource.
    """
    session = boto3.Session()

    # Unregister the default Serializer
    session.events.unregister(
        event_name="before-parameter-build.dynamodb",
        unique_id="dynamodb-attr-value-input",
    )

    # Unregister the default Deserializer
    session.events.unregister(
        event_name="after-call.dynamodb",
        unique_id="dynamodb-attr-value-output",
    )

    injector = CustomTransformationInjector()

    # Register our own serializer
    session.events.register(
        "before-parameter-build.dynamodb",
        injector.inject_attribute_value_input,
        unique_id="dynamodb-attr-value-input",
    )

    # Register our own deserializer
    session.events.register(
        "after-call.dynamodb",
        injector.inject_attribute_value_output,
        unique_id="dynamodb-attr-value-output",
    )

    return session
```

Finally, we're done. We've replaced the original (de)serializers with our own ones. On the one hand, it's cool that we can do that without _modifying_ the existing implementation. On the other hand, it wasn't very intuitive how to get here. Anyhow, let's see if any of this actually works.

To do that, I wrote three tests that can either run against an in-memory mock of DynamoDB courtesy of moto (see also: [Getting started with testing DynamoDB code in Python](https://www.tecracer.com/blog/2022/05/getting-started-with-testing-dynamodb-code-in-python.html)) or a real table in AWS. If you install the requirements and run the code from the [Github Repository](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/dynamodb-more-types), you should see something like this:

```text
$ python test_type_serializer.py
==================== test session starts ====================
platform darwin -- Python 3.9.17, pytest-7.1.2, pluggy-1.0.0
rootdir: /.../projects/python
plugins: Faker-13.3.3, freezegun-0.4.2, anyio-3.5.0, dash-2.4.1
collected 3 items

tests/test_type_serializer.py ...                     [100%]

===================== 3 passed in 1.21s =====================
```

Here's one of the test cases that also illustrates how you can use the modified boto session:

```python
def test_that_we_can_store_and_read_floats_in_ddb(table_fixture):
    """
    Test that we can store an item with a float object
    in it and retrieve it at a later time with the attribute
    still being of type float.
    """
    # Arrange
    tbl = build_boto_session().resource("dynamodb").Table("data")
    item_with_float_data = {
        "PK": "item_with",
        "SK": "float_data",
        "binary_data": "Hello World!".encode(),  # To ensure we don't break binary storage
        "pi": 3.14159265359,  # yes, there's more
    }

    # Act
    tbl.put_item(Item=item_with_float_data)

    item_from_ddb = tbl.get_item(Key={"PK": "item_with", "SK": "float_data"})["Item"]

    # Assert
    assert item_with_float_data == item_from_ddb
    assert isinstance(item_from_ddb["pi"], float)
```

If you want to use this in your own code, you just need to copy the `CustomSerializer`, `CustomDeserializer`, `CustomTransformationInjector`, and `build_boto_session`, including the imports and then create your resources as shown in the test case.

You can now use the custom (de)serializers to store other complex types in DynamoDB should you wish to extend it. Just be aware that this deserialization will only work if the requester also uses the deserializer. Otherwise, it will just look like random binary information to them.

In this blog post, we learned how to teach the table resource in boto3 to handle more data types by writing custom serializers and deserializers and injecting them into the SDK's data flow.

Thank you for reading this far, and hopefully, you learned something new.

&mdash; Maurice

---

Title Photo by [Jamie Matociños](https://unsplash.com/@jamievalmat?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText) on [Unsplash](https://unsplash.com/photos/rW00Wu_CeYA?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText)
