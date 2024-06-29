---
title: "What is Amazon Ion, and how can I read and write it in Python?"
author: "Maurice Borgmeier"
date: 2022-04-26
toc: false
draft: false
image: "img/2022/04/ion_title.png"
thumbnail: "img/2022/04/ion_title.png"
categories: ["aws"]
tags: ["level-300", "amazon-ion", "python"]

---

[Amazon Ion](https://amzn.github.io/ion-docs/) is a data serialization format that was open-sourced by Amazon in 2016 and is used internally at the company. Over time it has also been introduced into some AWS services and is the data format that services like the Quantum Ledger Database (QLDB) use. It has also started to appear in more commonly used services, so I think it's worth taking a closer look at. This article will explain what Ion is, its benefits, and how you can use it in Python.

You may have noticed that Ion popped up every once in a while in the AWS news feed in the last few years as being adopted or supported in different AWS services. I've collected a few services and features that use it here:

- [Redshift Spectrum can read Ion files since 2018](https://aws.amazon.com/about-aws/whats-new/2018/03/amazon-redshift-spectrum-now-supports-scalar-json-and-ion-data-types/)
- [DynamoDB can export tables in Ion format](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DataExport.Output.html)
- [The Quantum Ledger Database is built on top of the Ion format](https://docs.aws.amazon.com/qldb/latest/developerguide/ion.html)
- [Athena can now read and write Ion files](https://docs.aws.amazon.com/athena/latest/ug/ion-serde.html) ([since April 2022](https://aws.amazon.com/about-aws/whats-new/2022/04/amazon-athena-querying-ion-data/))
- [PartiQL can work with Ion data](https://partiql.org/faqs.html#why-do-you-choose-ion-to-extend-SQLs-type-system-over)

Ion is a data [serialization](https://en.wikipedia.org/wiki/Serialization) format. This means data structures can be transformed into Ion to store or transmit them. Later, they can be turned into data structures again. That in itself is nothing spectacular. There are many other data serialization formats, like CSV, JSON, Parquet, or Protobuf. Each of these formats has benefits and drawbacks and is designed for different use cases.

The [documentation](https://amzn.github.io/ion-docs/guides/why.html) on Ion has a section that explains why the data format is useful, and I encourage you to check it out because you may learn a thing or two about designing serialization formats and data types. Here are my key takeaways:

- Ion is a superset of JSON, which means any JSON document is a valid Ion document (but not vice-versa). This means you will most likely have already worked with Ion data.
- Comments are supported in the text representation of data.
- You can store and transmit Ion in both text and binary formats, making it possible to start with human-readable text and switch to an optimized binary format later.
- It has native support for timestamps of arbitrary precision, which is superior to storing timestamps as text.
- It supports [BLOB](https://en.wikipedia.org/wiki/Binary_large_object) data types and symbol tables to efficiently store commonly occurring terms.
- It supports the [decimal](https://docs.python.org/3/library/decimal.html) data type, making it suitable for applications that require math that works for humans and not computers.
- It's optimized for read-heavy workloads as most files are read more frequently than written and comes with a few tricks to speed up this process.

Here, we can see an example of Ion data that I've borrowed from the [docs](https://amzn.github.io/ion-docs/). This shows the text representation of an Ion file. The binary version isn't human-readable.

```yaml
/* Ion supports comments. */
// Here is a struct that is similar to a JSON object
{
  // Field names don't always have to be quoted
  name: "Fido",

  // This is an integer with a 'years' annotation
  age: years::4,

  // This is a timestamp with day precision
  birthday: 2012-03-01T,

  // Here is a list, which is like a JSON array
  toys: [
    // These are symbol values, which are like strings,
    // but get encoded as integers in binary
    ball,
    rope,
  ],

  // This is a decimal -- a base-10 floating point value
  weight: pounds::41.2,

  // Here is a blob -- binary data, which is
  // base64-encoded in Ion text encoding
  buzz: {{VG8gaW5maW5pdHkuLi4gYW5kIGJleW9uZCE=}},
}
```

Let's see it in action! To see some of the features that Ion offers, we will use data about the weekly fuel prices in Italy supplied by the [Ministry of Ecological Transition](https://dgsaie.mise.gov.it/open-data). First, we'll download the CSV data and then use pandas to prepare it. You can download the file using `wget` like this:

```terminal
wget "https://dgsaie.mise.gov.it/open_data_export.php?export-id=1&export-type=csv" -O "weekly_fuel_prices.csv"
```

Here is an excerpt from the data.

```csv
DATA_RILEVAZIONE,BENZINA,GASOLIO_AUTO,GPL,GASOLIO_RISCALDAMENTO,O.C._FLUIDO_BTZ,O.C._DENSO_BTZ
2005-01-03,1115.75,1018.28,552.5,948.5,553.25,229.52
2005-01-10,1088,1004.39,552.57,947.94,554.22,238.37
2005-01-17,1088.14,1004.31,551.88,952.42,562.78,245.89
```

As you can see, we have a date column and then a few decimal columns with the prices for the different fuels. Using pandas, we'll load the CSV data and transform it into data types that are more useful for calculations.

```python
import decimal
import pandas as pd

# Read as string to avoid loss of precision with floats
df = pd.read_csv("weekly_fuel_prices.csv", dtype=str)

# Rename to the english names
df = df.rename(columns={
    "DATA_RILEVAZIONE": "date",
    "BENZINA": "Euro-Super 95",
    "GASOLIO_AUTO": "Automotive Gas Oil",
    "GPL": "LPG",
    "GASOLIO_RISCALDAMENTO": "Heating Gas Oil",
    "O.C._FLUIDO_BTZ": "Residual Fuel Oil",
    "O.C._DENSO_BTZ": "Heavy Fuel Oil",
})
# Parse Date
df["date"] = pd.to_datetime(df.date)

# Convert the prices to decimals for precise arithmetic
from decimal import Decimal
for col in df.columns[1:]:
    df[col] = df[col].apply(Decimal)
```

Afterward, the data in the table looks as follows. We can't see the data types here, but aside from the date column, which is a timestamp, all other columns have the decimal data type.

![Data after conversion](/img/2022/04/ion_sample_data.png)

Now it's time to transform this data into Ion using the `amazon.ion` [package](https://pypi.org/project/amazon.ion/). First, we'll take a look at how this data would be saved in the text representation:

```python
import amazon.ion.simpleion as ion

# Convert to dataframe to list of dictionaries
list_of_values = df.to_dict("records")

# Only the first two rows in text format
print(ion.dumps(list_of_values[:1], binary=False, indent="  "))
```

The output of this code is the first row in Ion format. Here, we can see a few of the features of the format in action. The date column is a native timestamp in [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) format. The column names have been encoded with quotation marks where needed, and the decimal numbers have been stored as decimals without loss of precision.

```text
$ion_1_0
[
  {
    date: 2005-01-03T00:00:00.000000-00:00,
    'Euro-Super 95': 1115.75,
    'Automotive Gas Oil': 1018.28,
    LPG: 552.5,
    'Heating Gas Oil': 948.5,
    'Residual Fuel Oil': 553.25,
    'Heavy Fuel Oil': 229.52
  }
]
```

We could even add comments here `// using the leading double slash` or the `/* multiline comment syntax familiar from other languages*/`. This textual representation is verbose, though. When we store the data in binary form, we get more benefits. The interface of the library is fairly similar to the `json` module in the standard library, with the difference that it handles date-time and decimal data well.

```python
import amazon.ion.simpleion as ion

def save_as_ion(dict_or_list, file_name):
    with open(file_name, "wb") as file_handle:
        ion.dump(dict_or_list, file_handle)

save_as_ion(list_of_values, "weekly_fuel_prices_text.ion")
```

There is no significant size difference from the original CSV data for this tiny dataset. The original CSV is 47755 bytes in size, and the Ion file is 41777 bytes. The main gain here is type safety, which we can see when we deserialize, i.e., read the data again.

```python
import amazon.ion.simpleion as ion

def read_ion(file_name):
    with open(file_name, "rb") as file_handle:
        return ion.load(file_handle)

data_from_ion = read_ion("weekly_fuel_prices.ion")
```

If we inspect the data, we can see that it successfully read the timestamp and decimal information and converted it to the corresponding Python types.

![Ion Deserialization Result](/img/2022/04/ion_deserialize_result.png)

The `amazon.ion.simple_types.IonPyTimestamp` is actually a [subclass](https://ion-python.readthedocs.io/en/latest/amazon.ion.html#amazon.ion.simple_types.IonPyTimestamp) of the native [datetime class](https://docs.python.org/3/library/datetime.html#datetime-objects) in the standard library (albeit through another [level](https://ion-python.readthedocs.io/en/latest/amazon.ion.html#amazon.ion.core.Timestamp) of inheritance), so it can be used interchangeably with it. The same is true for the other types the `load` method returns. 

In this post, we've explored the data serialization format Ion and how you can read and write it from Python. Hopefully, you learned something. For any questions, feedback, or concerns, feel free to reach out to me via the social media channels in my bio.

&mdash; Maurice

