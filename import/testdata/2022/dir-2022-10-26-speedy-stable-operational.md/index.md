---
title: "10 years and one month: speed up website hosting on AWS in four steps"
author: "Gernot Glawe"
date: 2022-10-26
draft: false
image: "img/2022/09/100.png"
thumbnail: "img/2022/09/99.png"
toc: true
keywords:
    - architecture
    - cloudfront
    - amplify
    - s3
tags:
    - level-200
    - cloudfront
    - amplify
    - s3


categories: [aws]

---

There is no (milli)second chance for the first impression. Many websites today mess this up badly. When I need to wait 10 seconds for the content to load - I am out. What about you? I show you how to optimize the speed in four steps with S3, CloudFront and Amplify.



<!--more-->

![steps](/img/2022/10/cloudfront/steps.png)

I will use these different technologies to speed up:

![Overview](/img/2022/10/cloudfront/overview.png)



## Step one: File Size

I am not going deep into that topic because there are thousands of articles about that. But the first step in speeding up your page content begins with the content itself. Resize the images to a reasonable size; you do not need 4k resolution on a website. Then choose the proper format for your pictures, and weigh size versus quality.

As an example, I show you a picture which could be a symbol for your not-optimized website straight out of the mobile cam 🤓:

![rubbish](/img/2022/10/cloudfront/IMG_6202.jpeg)

The resolution of this picture is 4032 × 3024, which results in a size of 2.4 Mb—changing the resolution with the mac preview tool to 1008 × 756 results in 233 Kb, which loads much faster.


![rubbish small](/img/2022/10/cloudfront/IMG_6202-small.jpeg)

Doing the resize with `ffmpeg`, you even get down to 88k.

```bash
ffmpeg -i IMG_6202.jpeg -vf scale="1080:-1" IMG_6202-ffmpeg.jpg
```

![rubbish small](/img/2022/10/cloudfront/IMG_6202-ffmpeg.jpg)


```bash
ls -lkh IMG*
-rw-r--r--  1 gglawe  staff    88K 14 Okt 09:57 IMG_6202-ffmpeg.jpg
-rw-r--r--@ 1 gglawe  staff   233K 14 Okt 09:50 IMG_6202-small.jpeg
-rw-r--r--@ 1 gglawe  staff   2,3M 14 Okt 09:36 IMG_6202.jpeg
```

Decide for yourself about the quality loss of the picture - I can't see any difference.

## Step two: Static hosting

[aws-blog.de](https://aws-blog.com/) 

AWSblog.de is the blogging site of tecRacer. Co-founder Sven Ramuschkat started the blog **ten years ago**. See the [First Blog Entry 2012](https://aws-blog.com/2012/03/willkommen-beim-aws-blog.html). As a german company, we began in Deutsch and switched to English later. A few years later, I joined in. Now we have a steady team of bloggers.

![celebrate](/img/2022/10/cloudfront/celebrate.jpg)
**Celebrate 10 Years tecRacer AWS Blog**

It started as a WordPress blog because that is very convenient for authors. But there is a downside. WordPress generates the pages on demand, so with each request, there is additional time to render the page on the server.

### Migration from WordPress to S3

In the year 2015, our WordPress installation, running on an EC2 instance with a MySQL database, became unstable. Also, the content became more centred towards code. Therefore we decided to migrate the blog to [goHugo](https://gohugo.io/) and got some advantages:

#### 1) Stable static pages on S3

Static pages on S3 are stable, S3 will not crash. Never happened.



#### 2) Simpler Code integration

In hugo you can add code snippets like this shell command:


        ```bash
        ffmpeg -i IMG_6202.jpeg -vf scale="1080:-1" IMG_6202-ffmpeg.jpg
        ```

Which hugo will render as:

```bash
ffmpeg -i IMG_6202.jpeg -vf scale="1080:-1" IMG_6202-ffmpeg.jpg
```

#### 3) Page Speed

System    |Google Speed index
--- | ---
Wordpress  | ![wordpress-speed](/img/2019/05/wordpress.png)
Hugo      |  ![hugo-speed](/img/2019/05/334c9cdf.png)

You see that the [Google page speed index](https://pagespeed.web.dev/report?url=https%3A%2F%2Faws-blog.com%2F&form_factor=desktop) has much improved back then!

### HUGO Deployment

For the following steps, I will use [https://www.go-on-aws.com/](https://www.go-on-aws.com/) to measure caching effects.

Please refer to the HUGO documentation for details of creating a [HUGO](https://gohugo.io/) website. The deployment process has two commands:

- Generate static pages with the base url:

```bash
hugo  --baseURL {{.baseurl}}
```

- Copy the generated files to S3:

```bash
aws s3 sync  . s3://{{.bucket}}/ --delete --exclude ".git/*"
```

## Step three: regional caching

With [AWS Amplify Hosting](https://docs.aws.amazon.com/amplify/latest/userguide/welcome.html) you can really simplify website hosting. You don't have to manage storage, DNS and CDN caching (3,4). 

![Amplify Hosting](/img/2022/10/cloudfront/amplify-1.png)



You connect Amplify Hosting to the branch of your repository (1). Amplify Hosting (2) generates a HUGO configuration for `amplify.yml`:

```yaml
version: 1
frontend:
  phases:
    build:
      commands:
        - hugo
  artifacts:
    baseDirectory: public
    files:
      - '**/*'
  cache:
    paths: []
```

To add caching, https certificates and domain linking, you only add a R53 domain:

![amplify dns](/img/2022/10/cloudfront/amplify-domain.jpg)

Amplify will add a *hidden* certificate and a *hidden* cloudfront (3). You will not see a CloudFront distribution in your account. But Amplify adds this entry to the R53 zone:

Record | Type | Routing | Value
:------|:-----|:--------|:-----
www.go-on-aws.com | CNAME | Simple | d3ervqc58400yu.cloudfront.net


But why do I call this regional caching? Wait for the measurements to find out!

### Deploy Amplify

You just push to your repository (1) and amplify will start the build in pipeline:

![cicd](/img/2022/10/cloudfront/cicd.jpg)

## Step four: global caching

![cicd](/img/2022/10/cloudfront/cloudfront-1.png)

Now I create all AWS resources by myself. See the post [Building a static website with Hugo and the CDK](https://aws-blog.com/2020/05/building-a-static-website-with-hugo-and-the-cdk.html) from [Maurice](https://aws-blog.com/authors/maurice-borgmeier.html) for details how to do this with automatically with IaC.

This simplified architecture has 3 steps:

1) Generate pages locally with HUGO:

```bash
hugo
```

2) Synchronize local html  pages and images with s3 bucket:

```bash
cd public && aws s3 sync . s3://build.go-on-aws.com --delete --exclude ".git/*" --profile go-on-aws
```

3) CloudFront

Invalidate CloudFront Distribution:

```bash
aws cloudfront create-invalidation --distribution-id "{{.distid}}"   --path "/*"
```

If you do not invalidate, old files will be valid until the TTL (time to live) expires. So changes especially in `index.html` are not displayed immediately.


## Measure the results

Now I measure the page speeds of the different approaches. Because I want to test from *different locations*, I do not use Google pagespeed, but [pingdom](https://www.pingdom.com/). With pingdom you can choose different geographic locations for the test.

I performed the measurement over a period of **one month**. In this way, statistical outliers are smoothed out on average.

### Speed and architecture considerations

#### Location of the requester

Your customers access your site from several locations. So the measurement should be from several locations also.

#### Cloud-Provider of the requester

The sites are hosted on AWS. So when the requester site is also located on AWS, the traffic would only go through the AWS network. This type of measurement would give unrealistic values.

#### What is measured

Each measuring site has its own mix of values. So when you compare values, use *one* metric only. Comparing google page speed with pingdom won't get you anywhere...

#### Request several times

The network load on the internet varies over time. So you have to measure several times to get reasonable values.

#### Do not take first value

We want to test the caches, which might be filled on the first request. Therefore, the first request might take longer than the average.

### Stability

How often is the website down or how vulnerable to attacks is it?

### Easy of operation

What do I have to to to get the data to the server.

## Results

### Data

Type/Location | US    |   Asia  |   Europe
:------------ | :---- | :------ | :--------
Amplify | 1.24 | 2.18 |  0.738
S3 | 1.85 | 3.82 | 0.719
S3+CloudFront | 0.591 | 0.568 | 0.537

The original results from pingdom:

![pingdom data](/img/2022/10/cloudfront/2022-09-22_15-46-42.png)

Here you see the different URLs which I used for the different architectures.

### Overview

![Load time](/img/2022/10/cloudfront/median-load-time.png)

Smaller values are better, because of the faster load time.

### Interpretation

*CloudFront is the fastest*

Because the S3 bucket is located in Europe in region `eu-central-1`, the load time is below 1 second. With CloudFront, you squeeze another 25% speed advantage out of the bucket.

So if you only have customers in Germany, CloudFront always *gives you speed advantage compared to S3*. If you look at the multiple locations of the [Edge locations](https://aws.amazon.com/cloudfront/features/?whats-new-cloudfront.sort-by=item.additionalFields.postDateTime&whats-new-cloudfront.sort-order=desc), you see that, e.g. for Germany locations are not only in Frankfurt, but also in Berlin, Hamburg, Düsseldorf, Munich. So this brings your data nearer to the customer.

*Amplify Hosting behaviour*

Amplify is faster than S3, but with the region `eu-central-1` of the Amplify Hosting service the data implies that there is a regional cache reserved for amplify. In Europe CloudFront only is 0.019 seconds faster, which is negligible.
But outside the home region, in US CloudFront is about *100% faster than Amplify Hosting*!



## Conclusion

I have shown you how the different methods of Load-time optimization work.
Amplify really is a managed service which helps you with management. So if you only have customers in Europe, Amplify is easy and OK from the load speed.

What surprised me is the significant difference between Amplify Hosting and CloudFront speed outside the home region.
With the Amplify DNS entry, it becomes clear that Amplify does use CloudFront. But this seems to have only a small effect outside the home region.

So I would recommend going with S3/CloudFront whenever possible. 

If you need consulting for your CloudFront project, don't hesitate to get in touch with the sponsor of this blog, [tecRacer](https://www.tecracer.com/kontakt/).

For more AWS development stuff, follow me on twitter [@megaproaktiv](https://twitter.com/megaproaktiv)

## See also 

- [First Blog Entry 2012](https://aws-blog.com/2012/03/willkommen-beim-aws-blog.html)
- [Velocity Templates for Cloudformation](https://aws-blog.com/2017/01/velocity-for-complex-cloudformation-templates.html)
- [Blog migration to static website 2019](https://aws-blog.com/2019/05/serverless-blog-migration.html)



## Thanks to

- Celebration image generated by DALL·E

  