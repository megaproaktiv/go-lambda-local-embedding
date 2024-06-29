---
title: "High-level AWS cost benchmark in European regions (including new Zurich region!)"
author: "Benjamin Wagner"
date: 2022-11-09
toc: false
draft: false
image: "img/2022/11/aws_eu_regions_cost_comparison_heat_map_banner.png"
thumbnail: "img/2022/11/aws_eu_regions_cost_comparison_heat_map_banner.png"
tags: ["Costs"]
summary: | 
    As the new AWS region in Zurich has just been launched, I am wondering about the costs of AWS infrastructure in that region. So I decided to perform a small benchmark to get a better feeling about this.
---

<!-- As the new AWS region in Zurich has just been launched, I am wondering about the costs of AWS infrastructure in that region. So I decided to perform a small benchmark to get a better feeling about this. -->

# Welcome Switzerland!

As the new [AWS region in Zurich has just been launched](https://aws.amazon.com/blogs/aws/a-new-aws-region-opens-in-switzerland/), I am wondering about the costs of AWS infrastructure in that region. So I decided to perform a small benchmark to get a better feeling about this.

# Approach

My benchmark consists of a set of AWS services that I click together in the [AWS Pricing Calculator](https://calculator.aws/#/). I used a number of AWS services for that in order to see if there are remarkable deviations between services. With these same configurations, I used the Pricing Calculator once for each AWS region in Europe: Ireland, London, Paris, Frankfurt, Stockholm, Milan and Zurich. Here's one example: [AWS Pricing Calculator Example](https://calculator.aws/#/estimate?id=e1d7db910134b94c8e47f419eb7767aa6da60eed)

![Testing Configurations](/img/2022/11/aws_eu_regions_cost_comparison_configuration.png)

# Results

I assembled the results for each calculator in an Excel sheet. As we can see, there are no significant differences between the AWS services: A region that is more expensive than another region is also more expensive for each individual service that is included in the benchmark. The only exception is the data transfer cost as it has a globally unified price of 0.09 USD per GB. Slightly surprising is that Stockholm is the cheapest AWS region in Europe, and as expected, Zurich is a bit more expensive compared to all other regions.

![Benchmark Results](/img/2022/11/aws_eu_regions_cost_comparison_results.png)

Furthermore, I created a heatmap that shows the effect on the AWS costs when potentially moving from one region (y axis / lefthand side) to another (x axis / at the top). For example, migrating from eu-central-1 (Frankfurt) to eu-central-2 (Zurich) will increase your AWS bill by around 9%, and migrating from eu-west-1 (Ireland) to eu-north-1 (Stockholm) will reduce the bill be around 5%.

![Benchmark Results](/img/2022/11/aws_eu_regions_cost_comparison_heat_map.png)

# Conclusion

The results of the benchmark provide a ballpark figure of differences in AWS pricing in different regions. Of course, valid business cases should be based on more precise calculations. But at least, the heatmap may help to narrow down to just a few AWS regions to compare with each other. Keep in mind that the selection of the right AWS region not only depends on costs but also on compliance requirements, latency for your users as well as service and feature richness.

The benchmark also shows that the relative prices of AWS services within a region are more or less the same: In costly regions, all individual services are costly, and in cheaper AWS regions, all individual services are cheaper. Note that we included only a number of AWS services and with other services, there might be more significant deviations (however I don't really expect that).

Lastly, one big takeaway for me is that in fact the Stockholm region as the lowest prices in Europe, and not Ireland as I expected.