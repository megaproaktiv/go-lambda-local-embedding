---
title: "HIVE_CURSOR_ERROR in Athena when reading parquet files written by pandas"
author: "Maurice Borgmeier"
date: 2023-08-07
toc: false
draft: false
image: "img/2023/08/athena_error.png"
thumbnail: "img/2023/08/athena_error.png"
categories: ["aws"]
tags: ["level-300", "athena"]
summary: |
  In a recent project, a colleague asked me to look at a `HIVE_CURSOR_ERROR` in Athena that they weren't able to get rid of. Since the error message was not incredibly helpful and the way this error appeared is not that uncommon, I thought writing this may help you, dear reader, and future me when I inevitably forget about it again.
---

In a recent project, a colleague asked me to look at a `HIVE_CURSOR_ERROR` in Athena that they weren't able to get rid of. Since the error message was not incredibly helpful and the way this error appeared is not that uncommon, I thought writing this may help you, dear reader, and future me when I inevitably forget about it again.

![Error Message](/img/2023/08/athena_error.png)

The error message is as follows:

> HIVE_CURSOR_ERROR: Failed to read Parquet file: s3://bucket/table/file.parquet

It doesn't tell us much, and neither does the AWS documentation at the time of writing this. Let's figure out where it's coming from.

## Diagnosis

The original query was a simple `select * from data_table limit 10`, nothing fancy. This was a strong indicator that something in the table configuration didn't match the underlying data in the parquet file.

As a first step, we tried reading individual columns like this:

```sql
select column_a
  from data_table
  limit 10
```

Depending on which column we selected, the query either returned the same error or the expected result, which pointed to a mismatch between _some_ of the data types in the underlying parquet and table.

The next step was downloading and inspecting the parquet file indicated in the error message. There's a very useful [python tool](https://pypi.org/project/parquet-tools/) that we can leverage to gain access to the parquet metadata called `parquet-tools`.

```terminal
$ pip install parquet-tools
$ parquet-tools inspect my_suspicious.parquet
[...]
############ Column(column_a) ############
name: column_a
path: column_a
max_definition_level: 1
max_repetition_level: 0
physical_type: DOUBLE
logical_type: None
converted_type (legacy): NONE
compression: SNAPPY (space_saved: 0%)
[...]
```

The parquet tools inspect the metadata and show us the underlying data type of the columns. Here it turned out that the columns were of type `DOUBLE` in parquet, whereas the Glue Data Catalog described them as a `BIGINT`. That explains why Athena couldn't read it, although the error message could be improved.

This leaves the question of why there is a `DOUBLE` when the table defines a `BIGINT`. In this case, `BIGINT` was the correct data type because the column is supposed to contain nullable integers, i.e., integers or null. So what happened here?

## Root Cause Analysis

The parquets are written through pandas/pyarrow, and the pandas library is causing our problem here. When the data is requested from an API, it's converted into a list of dictionaries. The attribute/column either contains the integer or null (None in Python). In order to turn this into a parquet file, we create a Pandas Dataframe from our list of dictionaries.

Here, something interesting is happening. When pandas notices the None/null in the data, it replaces it with `numpy.NaN`, which has the datatype `float` since integers in the underlying Numpy array can't be null.

```python
>>> import numpy as np
>>> type(np.NaN)
<class 'float'>
```

When writing this to a dataframe, pandas must choose a type that fits all the data in the column. Since there is at least one floating point number in there, and you can represent integers as floats, it chooses the underlying double (more precise float) data type for the parquet.

This is how we can create the `my_suspicious.parquet` file:

```python
import pandas as pd

def main():
    data = {
        "column_a": [1, None, 3, 4],
    }

    frame = pd.DataFrame(data)

    frame.to_parquet("my_suspicious.parquet", index=False)

if __name__ == "__main__":
    main()
```

## Fixing it

Since the desire to have nullable integers is not that uncommon, the [pandas documentation](https://pandas.pydata.org/docs/user_guide/integer_na.html) offers some ways to work around this. We just need to cast the column to a special nullable Integer type like this:

```python
import pandas as pd

def main():
    data = {
        "column_a": [1, None, 3, 4],
    }

    frame = pd.DataFrame(data)
    frame.column_a = frame.column_a.astype(pd.Int64Dtype())

    frame.to_parquet("my_suspicious.parquet", index=False)

    # Assert that our column is now of type integer
    assert pd.api.types.is_integer_dtype(frame.column_a)


if __name__ == "__main__":
    main()
```

If we run the `parquet-tools` on the resulting file again, we can see that it now has the proper data type `INT64`, which is compatible with Athena's `BIGINT`.

```terminal
$ parquet-tools inspect my_suspicious.parquet                       [...]
############ Column(column_a) ############
name: column_a
path: column_a
max_definition_level: 1
max_repetition_level: 0
physical_type: INT64
logical_type: None
converted_type (legacy): NONE
compression: SNAPPY (space_saved: -1%)
```

Under the hood, pandas uses the `pandas.NA` value to represent the nulls, which allows it to store them in a numpy array.

## Conclusion

This `HIVE_CURSOR_ERROR` may indicate a mismatch between the table's expected data type and the parquet file's underlying data type. I showed you how to diagnose and fix this issue if it's caused by nullable integers in pandas.

&mdash; Maurice
