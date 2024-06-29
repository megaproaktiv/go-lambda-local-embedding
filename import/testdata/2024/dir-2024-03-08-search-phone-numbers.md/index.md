---
title: "Find It Fast: Streamline Phone Number Searches with OpenSearch."
author: "Alexey Vidanov"
date: 2024-03-08
toc: true
draft: false
image: "img/2024/02/opensearch-phone.png"
thumbnail: "img/2024/02/opensearch-phone.png"
categories: ["aws"]
tags: ["aws", "opensearch", "onelogin", "level-400", "enterprise search", "search", "index"]
---

This guide empowers you to optimize OpenSearch for lightning-fast and accurate phone number searches. Frustration-free experiences are key for your customers, and by leveraging edge ngrams and custom analyzers, you can empower OpenSearch to efficiently handle even large datasets. 

<!--more-->

**Challenge:** Traditional relational databases often struggle with the demands of modern applications, particularly those with dynamic search features like "search as you type." This can lead to performance bottlenecks, increased costs, and limitations in functionality. This was the case for our client, a start up company, who develops a cloud-native restaurant management system built on AWS.

**Solution:** In collaboration with our client, we developed and implemented a one-field comprehensive search solution on Amazon OpenSearch Service, including phone number search. This collaborative effort helped bypass relational database limitations, resulting in faster, more accurate queries and significantly enhanced dynamic search functionality.

We'll delve into data preparation, phone number indexing optimization, and index setup, all geared towards enhancing search speed and precision. Remember, customization is crucial to aligning search functionality with your specific data needs.

## Prerequisites

To proceed with the instructions in this blog and successfully implement real-time phone number searches using OpenSearch, ensure you have the following prerequisites ready:

- **Amazon Web Services (AWS) Account:** Access to an AWS account is necessary to utilize Amazon OpenSearch Service. If you don't have one, you can sign up for an AWS account [here](https://aws.amazon.com/).
- **Basic Understanding of OpenSearch:** Familiarity with the fundamentals of OpenSearch, including its architecture and core concepts, will greatly aid in following the instructions provided.
- **Amazon OpenSearch Service Documentation:** Review the [Amazon OpenSearch Service documentation](https://docs.aws.amazon.com/opensearch-service/) for detailed guidance on installation, configuration, and management of your OpenSearch cluster.
- **SDK Installation:** Depending on your preferred programming language, ensure the appropriate SDK is installed on your development machine. This will facilitate communication with the OpenSearch cluster through your application.
- **OpenSearch Dashboards:** While optional, having OpenSearch Dashboards set up can be beneficial for visualizing your data and testing your search queries in a more user-friendly environment.
- **Sample Data Preparation:** To effectively index phone number data, it's essential to have a set of sample data ready. For your convenience, this blog includes [dummy data](https://github.com/vidanov/tecracer-blog-projects/blob/main/opensearch_phone_numbers/opensearch_customers_dummy.txt), which we generated using [Python's Faker library](https://faker.readthedocs.io/en/master/). While this dataset serves as a good starting point, we recommend utilizing your own data for more accurate and realistic testing and implementation purposes.

Feel free to ask for more details on any topic mentioned. We're here to offer additional clarification and support.

## 1. Prepare your data

When storing phone numbers in the OpenSearch index, it's advisable to utilize two fields: one for the country code and one for the number itself. The method for achieving this depends on your setup. However, it's generally straightforward to normalize the phone numbers and separate the country code from the number using libraries like phonenumbers for Python and libphonenumber-js for Node.js and PHP.

```python
>>> import phonenumbers
>>> x = phonenumbers.parse("+442083661177", None)
>>> print(x)
Country Code: 44 National Number: 2083661177 Leading Zero: False
```

**References:**

- Python: [phonenumbers](https://pypi.org/project/phonenumbers/)

- Node.js: [libphonenumber-js](https://github.com/catamphetamine/libphonenumber-js)

- PHP: [libphonenumber-for-php](https://github.com/giggsey/libphonenumber-for-php)

  

##  2. Optimizing Phone Number Indexing with Edge Ngrams

Before indexing data, it's crucial to establish an appropriate mapping for phone number fields. 

### The Edge Ngram Approach

For enabling a dynamic "search as you type" functionality, the implementation of edge ngrams is preferred. This technique differs significantly from traditional ngrams by generating tokens exclusively from the beginning of the phone number. This method ensures a more targeted and efficient search process, avoiding the pitfalls associated with full-length ngram sliding.

### The Limitations of Regular Ngrams

Regular ngrams, despite their comprehensive coverage, introduce several drawbacks:

- **Increased Noise and Index Size**: They produce an extensive array of tokens, capturing every possible character combination within the phone number. This not only leads to irrelevant search results but also escalates the index size considerably.
- **Performance and Storage Concerns**: An experiment with a simple index demonstrated that an ngrams-based index was three times larger than one without any grams. In stark contrast, edge ngrams required only 20% additional storage, highlighting the inefficiency of regular ngrams.

### Advantages of Edge Ngrams

Edge ngrams present a strategic alternative with multiple benefits:

- **Improved Search Relevance**: By concentrating on the initial segments of the phone number, they ensure that search results are closely aligned with the user's query, minimizing irrelevant suggestions.
- **Reduced Noise**: The generation of fewer tokens directly translates to cleaner search results, enhancing the suggestion quality.
- **Lower Storage Footprint**: When compared to regular ngrams, edge ngrams substantially decrease the index size, leading to improved efficiency.

### A Practical Illustration

Consider the phone number "+1234567890". While regular ngrams would produce tokens like "1", "23", "234", etc., edge ngrams generate "1", "12", "123", etc. In a scenario where a user types "123", both methods might return the correct phone number. However, edge ngrams eliminate the creation of irrelevant tokens such as "456" or "7890", thereby streamlining the search experience.

Edge ngrams are highly effective for "search as you type" features, indexing the start of each term to provide swift and accurate search results. This approach significantly enhances user experience by ensuring relevance, reducing noise, and minimizing storage requirements.

## 3. Improve Accuracy by Handling Incomplete Numbers

To address issues such as missing digits or incorrect country codes, it's beneficial to adopt a strategy that includes an additional subfield and a custom search analyzer. Concentrating on matching the last four digits significantly boosts the chances of accurately identifying the caller. This approach enables us to overlook initial errors, such as incorrect or absent country codes, thus improving the robustness and dependability of the identification process.

Please note: The number of digits focused on for matching can be adjusted up or down based on the database size and specific needs, allowing for optimized results.

## 4. Prepare your index

Below is an example of how to structure your index mapping to incorporate the previously discussed techniques for enhanced accuracy. To streamline this demonstration, we will omit the country code. It's assumed that there is a standardized input normalization process for phone numbers implemented within the code. This normalization process is consistently applied to both the stored data and incoming queries to ensure uniformity and improve match accuracy.

```json
PUT /my_phone_numbers_index
{
  "settings": {
    "analysis": {
      "tokenizer": {
        "edge_ngram_digits_tokenizer": {
          "type": "edge_ngram",
          "min_gram": "3",
          "max_gram": "20",
          "token_chars": [
            "digit"
          ]
        }
      },
      "filter": {
        "last_four_digits": {
          "type": "pattern_capture",
          "preserve_original": false,
          "patterns": [
            """(\d{4})$"""
          ]
        }
      },
      "analyzer": {
        "phone_analyzer": {
          "type": "custom",
          "tokenizer": "edge_ngram_digits_tokenizer"
        },
        "phone_search_analyzer": {
          "type": "keyword"
        },
        "last_four_digits_analyzer": {
          "type": "custom",
          "tokenizer": "keyword",
          "filter": [
            "last_four_digits"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "phone_number": {
        "type": "text",
        "analyzer": "phone_analyzer",
        "search_analyzer": "phone_search_analyzer",
        "fields": {
          "last_four_digits": {
            "type": "text",
            "analyzer": "last_four_digits_analyzer"
          }
        }
      }
    }
  }
}
```

### Settings

**Analysis Configuration:** This section configures the analysis process, which is how text is processed and indexed.

- **Tokenizer**
  - `edge_ngram_digits_tokenizer`: A custom tokenizer of type `edge_ngram` configured to create tokens from the input text by breaking it down into edge n-grams of digits only. This tokenizer will generate tokens of lengths ranging from 3 to 20 characters, focusing exclusively on digit characters. This is useful for partial matching of phone numbers.
- **Filter**
  - `last_four_digits`: A custom filter of type `pattern_capture` that captures the last four digits of the indexed phone numbers. It does not preserve the original token.
- **Analyzers**
  - `phone_analyzer`: A custom analyzer using the `edge_ngram_digits_tokenizer` for indexing phone numbers. This analyzer is suitable for indexing phone numbers in a way that supports searching for partial numbers.
  - `phone_search_analyzer`: A keyword analyzer used during the search phase, ensuring that the search input is treated as a single token. This is useful for exact matches.
  - `last_four_digits_analyzer`: A custom analyzer tailored for indexing the last four digits of phone numbers. It uses the `keyword` tokenizer along with the `last_four_digits` filter.

### Mappings

**Properties**

- `phone_number` : Defines how the`phone_number` field is indexed and searched.
  - Type is set to `text`, making it suitable to apply `edge_ngram` tokenizer.
  - Uses `phone_analyzer` for indexing, enabling partial matches on phone numbers.
  - Uses `phone_search_analyzer` for searching, optimizing for exact match searches.
  - Introduces a sub-field `last_four_digits` analyzed by `last_four_digits_analyzer`, specifically designed for searches focused on the last four digits of phone numbers.
  

## 5. Testing of the analyzers

To evaluate the functionality of your analyzer, execute the following commands. These will generate edge n-gram tokens based on the analyzer configuration.

For the initial test, use the following request to analyze the phone number "+19876543210" with the `phone_analyzer`:

```json
GET /my_phone_numbers_index/_analyze
{
  "analyzer": "phone_analyzer", 
  "text": ["+19876543210"]
}
```

This query will process the input text through the specified analyzer and return the generated tokens.

For testing the `last_four_digits_analyzer`, which is designed to extract the last four digits of a phone number, input the same phone number as follows:

```json
GET /my_phone_numbers_index/_analyze
{
  "analyzer": "last_four_digits_analyzer", 
  "text": ["+19876543210"]
}
```

This request will result in the analyzer isolating and returning "3210" as the token, demonstrating the focused functionality of extracting the last four digits from the provided phone number.

## 6. Adding some phone numbers

You can start with the  [dummy data](https://github.com/vidanov/tecracer-blog-projects/blob/main/opensearch_phone_numbers/opensearch_customers_dummy.txt) or just put some numbers using an API requrst like this:

```json
POST /my_phone_numbers_index/_bulk
{ "index": { "_id": "2" }}
{ "phone_number": "+19876543210" }
{ "index": { "_id": "3" }}
{ "phone_number": "+11234567890" }
{ "index": { "_id": "4" }}
{ "phone_number": "+1231231234" }
```

## 7. Conduct a Search

To illustrate how to search within your index, consider the following example. This search operation aims to find documents in the `my_phone_numbers_index` that match the specified criteria for phone numbers.

```json
GET /my_phone_numbers_index/_search
{
  "query": {
    "bool": {
      "should": [
        {
          "match": {
            "phone_number": {
              "query": "29876543210",
              "boost": 2.0
            }
          }
        },
        {
          "match": {
            "phone_number.last_four_digits": "29876543210"
          }
        }
      ]
    }
  }
}
```

This query demonstrates the use of a `bool` query with `should` clauses, allowing for flexibility in matching documents. The first clause attempts to match the entire phone number with a boost factor of 2.0, giving it higher relevance if matched. The second clause seeks to match the last four digits of the phone number within a specific field designed to store these digits. This approach facilitates a search strategy, accommodating various ways a phone number might be queried or stored.

## What else you can do? The Indirect Benefits of Highlighting Query Matches

OpenSearch's highlighting feature goes beyond just presenting results. It visually emphasizes how your search terms match the data, making it easier for users to understand why a particular document is relevant. This improves comprehension and speeds up the decision-making process.

![image-20240223140832996](/img/2024/02/opensearch-search-highlight.png)

Here's how highlighting benefits your users:

- **Clarity:** Highlighted terms instantly reveal which parts of the document match the search query. This eliminates guesswork and saves users time spent sifting through irrelevant information.
- **Confidence:** By seeing the matched terms, users gain confidence in the search results' accuracy. This empowers them to make informed decisions based on the highlighted information.
- **Efficiency:** Highlighting streamlines the search process by directing users' attention to the most relevant parts of the document. This reduces time spent scanning through large amounts of text.

For a deeper understanding of how highlighting can be implemented within OpenSearch, visit [OpenSearch Highlighting Documentation](https://opensearch.org/docs/latest/search-plugins/searching-data/highlight/).

## In Conclusion

By leveraging edge ngrams and custom analyzers, OpenSearch empowers you to achieve blazing-fast and accurate phone number searches, even with massive datasets. This guide equipped you with the knowledge to prepare your data, optimize phone number indexing, and configure your index for superior search performance. Remember, customization is key! Tailor the search functionality to your specific data requirements to maximize effectiveness.

OpenSearch can give your business a significant edge by providing a frictionless search experience for your users. Don't be afraid to experiment and refine your approach to achieve the best possible results. If you have any questions or need further guidance on implementing a robust OpenSearch cluster and API, feel free to reach out. We're here to help you elevate your search capabilities beyond expectations. Happy searching!

— Alexey

