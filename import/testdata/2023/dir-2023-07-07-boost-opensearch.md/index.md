---
title: "Performance Boost: 10 Expert Tips for Optimizing Your Amazon OpenSearch Service Cluster"
author: "Alexey Vidanov"
date: 2023-07-08
toc: true
draft: false
image: "img/2023/07/boost.png"
thumbnail: "img/2023/07/boost.png"
categories: ["aws"]
tags:
  ["aws", "opensearch", "optimize", "data analytics", "level-400", "tips", "performance", "elasticsearch"]
---
By implementing these recommendations, you can maximize the potential of your Amazon OpenSearch Service domain, delivering an improved search experience while optimizing costs and maintaining security. Let's explore these expert tips to supercharge your OpenSearch cluster. 

<!--more-->

We've organized our 10 top tips for Amazon OpenSearch Service into five key categories: Hardware, Indexing, Monitoring, Sharding and Query Optimization. For each category, we've identified two key improvements based on common usage patterns. These tips should help you navigate OpenSearch more effectively. Let's get started.

## Hardware

![abstract hardware](/img/2023/07/hardware.png)

1. **Choose the right instance type** 
   Choosing the right instance type for your workload is crucial. For example, if you have a write-heavy workload, you should choose an instance type with high I/O performance.

   - Data Nodes with EBS: EBS can be a cost-effective option and works well for log data. It offers 3 IOPS (Input/Output Operations Per Second) per GB deployed.

   - Data Nodes with Instance Store (I3, r6gd Instances): Instance Store provides faster performance, with r6gd instances being the top performers. While I3 instances deliver slightly less performance, they offer the advantage of larger volumes.
   - For large workloads use dedicated master nodes: Using dedicated master nodes can help to improve the stability of your cluster.

2. **Start big** 
   Remember, it's easier to measure the excess capacity in an overpowered cluster than the deficit in an underpowered one, so it's recommended to start with a larger cluster than you think you need, then test and scale down to an efficient cluster that has the extra resources to ensure stable operations during periods of increased activity.

## Indexing

![abstract indexing big data](/img/2023/07/indexing.png)

3. **Use bulk ingest requests and employ multi-threading** 
   Bulk requests are more efficient than individual index requests. For example, a single thread can index 1000 small documents per second, but with bulk requests, it can index 100,000 to 250,000 documents per second. The bulk API is a powerful tool for indexing multiple documents in a single request, reducing the overhead of individual indexing requests. The optimal bulk size varies depending on the use case, but a good starting point is between 5-15MB.

   To enhance indexing throughput, employ multi-threading. This can be achieved using OpenSearch SDKs and libraries like `opensearch-py`. By creating 10-20 threads per node, you can significantly boost your indexing performance.

4. **Optimize** 
   Minimize frequent updates: To maximize efficiency in OpenSearch, minimize frequent updates to the same document. This prevents the accumulation of deleted documents and large segment sizes. Instead, collect necessary updates in your application and selectively transmit them to OpenSearch, reducing overhead and improving performance. As an example, when storing stock information in the index, it's recommended to represent it using levels (e.g., available, low, not available) instead of numerical values. This approach ensures efficient storage and retrieval of stock data in OpenSearch.

   Do not index everything: Disabling indexing for specific fields by setting `"index": false` in the field mapping can help optimize storage, improve indexing performance.

   Tune your _source field :

   - The `_source` field in OpenSearch is a special field that holds the original JSON object that was indexed. This field is automatically stored for each indexed document and is returned by default in search results.

   - The primary advantage of the `_source` field is that it allows you to access the original document directly from the search results. This can be particularly useful for debugging purposes or for performing partial updates to documents.

   - However, storing the `_source` field does increase storage requirements. Each indexed document essentially gets stored twice: once in the inverted index for searching and once in the `_source` field.

   - If your use case doesn't require accessing the original document in search results, you can disable storing the `_source` field to save storage space. This can be done by setting `"enabled": false` in the `_source` field mapping.

## Monitoring

![monitoring](/img/2023/07/monitoring.png)

5. **Use CloudWatch** 
   Monitoring tools like Amazon CloudWatch can be used to track indexing performance and identify bottlenecks. Enabling Slow Logs can save a lot of time. Set up [the recommended CloudWatch alarms for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cloudwatch-alarms.html)
6. **Profile queries**
   Profiling your OpenSearch queries can provide valuable insight into how your queries are being executed and where potential performance bottlenecks may be occurring. The Profile API in OpenSearch is a powerful tool for this purpose.
   - To use the Profile API, simply append `?profile=true` to your search queries. This will return a detailed breakdown of your query's execution, including information about how long each operation took and how the query was rewritten internally.
   - The output of the Profile API is divided into sections for each shard that participated in the response. Within each shard section, you'll find details about the query and aggregation trees.
   - The query tree shows how the query was executed across the inverted index, including the time taken by each term. The aggregation tree, on the other hand, shows how the aggregations were computed, including the time taken by each bucket.
   - By analyzing this information, you can identify which parts of your query are taking the most time and adjust them accordingly. This could involve changing the structure of your query, adjusting your index mappings, or modifying your OpenSearch cluster configuration.
   - Remember, profiling adds overhead to your queries, so it's best to use it sparingly and only in a testing or debugging environment.

## Sharding

![sharding](/img/2023/07/sharding.png)

7. **Find an optimal shard number and size:** The ideal shard size in OpenSearch is typically between 10GB and 50GB for workloads where search latency is a key performance objective, and 30-50GB for write-heavy workloads such as log analytics. Large shards can make it difficult for OpenSearch to recover from failure, but having too many small shards can cause performance issues and out of memory errors. 

   The number of primary shards for an index should be determined based on the amount of data you have and your expected data growth. A general guideline is to try to keep shard size between 10–30 GiB for workloads where search latency is a key performance objective, and 30–50 GiB for write-heavy workloads such as log analytics.

8. **Optimize shard locating:** Overallocating shards can lead to wasted resources. On a given node, have no more than 25 shards per GiB of Java heap. For example, an m5.large.search instance has a 4-GiB heap, so each node should have no more than 100 shards. At that shard count, each shard is roughly 5 GiB in size, which is well below the recommended size range.

## Search and Query Performance

9. **Use filters:** Filters are faster than queries because they don’t calculate relevance (_score). They simply include or exclude documents.
10. **Use search templates:** One effective way to boost your Amazon OpenSearch Service is by utilizing search templates. Search templates allow you to predefine and reuse complex search queries, reducing the processing time and improving search performance. 

**Additional Reading**

- [Operational best practices for Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/bp.html)

- [Get started with Amazon OpenSearch Service: T-shirt-size your domain](https://aws.amazon.com/blogs/big-data/get-started-with-amazon-opensearch-service-t-shirt-size-your-domain/)

- [Amazon OpenSearch Service FAQs](https://aws.amazon.com/opensearch-service/faqs)

- [Fine Dining for Indices](https://opensearch.org/blog/fine-dining-for-indices/)

- [How can I improve the indexing performance on my Amazon OpenSearch Service cluster?](https://repost.aws/knowledge-center/opensearch-indexing-performance) 

- [How do I resolve search or write rejections in OpenSearch Service?](https://repost.aws/knowledge-center/opensearch-resolve-429-error)

- [How to Improve your OpenSearch Indexing performance](https://opster.com/blogs/improve-opensearch-indexing-rate/)

- [How to Improve OpenSearch Search Performance](https://opster.com/blogs/improve-opensearch-search-performance/)

- [Sizing Amazon OpenSearch Service Domains](https://d1.awsstatic.com/events/reinvent/2021/Sizing_Amazon_OpenSearch_Service_domains_REPEAT_ANT402-R2.pdf)

  

Keep in mind, these tips and improvements are just the starting point. The ultimate effectiveness of their application depends on your specific scenario and use case. While we've focused on the areas of Hardware, Indexing, Sharding, Monitoring, and Optimization, there are many more facets to consider. For instance, security, which is a critical component we haven't delved into here. Remember, your OpenSearch Service should be as unique as your needs.

If you require assistance in optimizing your OpenSearch Service deployment, [tecRacer](https://www.tecracer.com/en/consulting/amazon-opensearch-service/), an Amazon OpenSearch Service Delivery Partner, is here to provide expert guidance. Our team of professionals specializes in designing, deploying and securely managing OpenSearch Service infrastructures tailored to individual needs. Whether you need support in selecting the right instance types, fine-tuning indexing strategies, monitoring performance, or optimizing search and query operations, tecRacer can provide the expertise you need.
