---
title: "Scheduling dev.to posts with Honeycode"
author: "Maurice Borgmeier"
date: 2022-03-08
toc: false
draft: false
image: "img/2022/03/hc_title.png"
thumbnail: "img/2022/03/hc_title.png"
categories: ["aws"]
tags: ["level-200", "honeycode", "no-code"]
summary: |
    I explain how I built an app that uses Honeycode and an API Gateway backed by a Lambda Function to schedule my blog posts on dev.to.
---

I publish my posts to our company blog at [aws-blog.de](https://aws-blog.de) and [dev.to](https://dev.to/mauricebrg). I like preparing them in advance and having them posted automatically. Our company blog handles this through a daily run of the CodePipeline that publishes blogs. Unfortunately, dev.to doesn't support something like that natively. Fortunately, they have an API, and since I wanted to try Amazon Honeycode anyway, I decided to build a scheduler myself.

If you're not familiar with Honeycode already, I [wrote an introduction](https://aws-blog.de/2022/03/amazon-honeycode-has-potential.html) to the service recently. I suggest you read it before we continue here. In a nutshell, Honeycode lets you build user interfaces around a spreadsheet-like database and enables you to trigger actions based on changes to data in that database. The user experience should be as simple as possible. My goal was to paste in the preview URL of a post, set a publication time, and then forget about it. Supporting more user accounts and overviewing scheduled and already published posts were side-goals.

I broke up my scheduler into two components, intending to put as much logic as possible into Honeycode. Honeycode contains the dev.to users and their API keys and the posts intended for publication. The user interface is responsible for wrangling the data and taking a peek into the upcoming schedule. The second component consists of an API gateway and a Lambda function. Its primary purpose is to take in a *dev.to* preview URL and an API key and publish the post. The Lambda function needs to do a few things to achieve that, but more on that later. First, let's focus on the data model in Honeycode.

![Data Model](/img/2022/03/devto_scheduler_data_model.png)

I created two tables in Honeycode, dev_to_posts, and dev_to_users. The _users_ table contains dev.to usernames and the respective API keys. There is just a single record in my case, but it would support more than that. Also, I wanted to configure this through the GUI and not redeploy the backend if something changes.

The other table contains the posts with a little bit of metadata. The _Author_ column links to the *dev_to_users* table, which sets up something like a foreign-key relationship. In the _Post URL_ column, you can find the preview URL from *dev.to*, _Publish At_ is a DateTime column with the publication date and time. The other two fields are not that interesting. There is a *Note* field for some comments and, in addition to that, a *Published Post URL* field that uses a formula to turn the preview URL from the *Post URL* column into the URL under which the post will be available. Here is the formula in case you plan to recreate that for yourself:

```excel
=LEFT([Post URL],SEARCH("-temp-slug",dev_to_posts[Post URL])-1)
```

Next, I used the Honeycode Builder to create a user interface around the data model. It contains an overview of posts for future publication, an option to add new posts given their preview URL, a list of the already published posts, and a way to manage the API key. Below you'll find some impressions of the implementation.

![App Screenshots](/img/2022/03/hc_screenshots.png)

I used filters in the source field to restrict the kinds of records shown in the *Scheduled Posts* and *Published Posts* lists. Creating the *New Scheduled Post* form was a breeze. I just selected the form component, dragged it on the *Scheduled Posts* screen, and selected the table where I wanted to add a record. That created the form and the required buttons, and I could customize it after the fact. The interface now allows me to interact with the data in my tables effortlessly.

```excel
// Scheduled Posts
=FILTER(dev_to_posts,"dev_to_posts[Publish At]>NOW() ORDER BY dev_to_posts[Publish At] ASC")

// Published Posts
=FILTER(dev_to_posts,"dev_to_posts[Publish At]<NOW() ORDER BY dev_to_posts[Publish At] ASC")
```

Triggering the publication backend whenever a new post needed to be published was the next task at hand. I set up an automation that fires whenever the *Publish At* time in a row of the *dev_to_posts* table is reached. Honeycode treats all times as UTC, so I had to add an offset for it to match my local timezone. The first action sends a notification letting me know that a post was published.

![Automation Trigger](/img/2022/03/hc_automation_trigger.png)

After sending a notification to me, the automation calls a webhook. Webhooks are HTTPS endpoints that can be invoked when events happen. Here, I configured it to send the Preview URL and the API Key of the author to that webhook. The webhook's back end uses this information to talk to the *dev.to* API and publish the post. In addition to the payload, I'm also setting an HTTP header with the API-Key for the backend.

![Webhook Config](/img/2022/03/hc_webhook.png)

The backend consists of a Lambda function behind an API-Gateway. This Lambda function extracts the post URL and API key from the event and uses the [*dev.to* API](https://developers.forem.com/api) to get the article ID of the post behind the preview URL. Assuming that works, it proceeds to publish the post with that ID. If you're curious, the complete code for the backend [is available on Github](https://github.com/MauriceBrg/dev-to-publisher).

```python
def lambda_handler(api_event, __unused):
    """
    Handles incoming requests from the API Gateway.
    """

    payload = json.loads(api_event["body"])

    assert "PostURL" in payload, "PostURL needs to be present"
    assert "ApiKey" in payload, "ApiKey needs to be present"

    post_url = payload["PostURL"]
    api_key = payload["ApiKey"]

    print(f"Got PostURL {post_url}")
    # We don't want to show the whole API key that wouldn't be great.
    print(f"Got API Key {api_key[0]}{ '*' * (len(api_key) - 2)}{api_key[-1]}")

    article_id = get_article_id_by_preview_url(
        preview_url=post_url,
        api_key=api_key
    )

    if article_id is not None:
        publish_article_by_id(article_id=article_id, api_key=api_key)

    return {}
```

## Summary

In this post, I showed you how I built a system to schedule blogs on *dev.to* using a combination of Honeycode, Lambda, and the API Gateway. I'm not the first to solve this problem. There is already an [Azure-based solution](https://dev.to/toddanglin/publishto-dev-scheduling-article-publishing-on-dev-to-3m0o) that works a bit differently and requires you to trust the provider with your API key and install a browser extension. You'll have to deploy fewer components than my solution requires on the plus side. For me, this was primarily an exercise to learn Honeycode and solve a problem of mine. Nevertheless, I'd appreciate it if *dev.to* adds a first-party solution to this problem. 

I hope you enjoyed reading this post. I'm happy to receive feedback via the social media channels in my bio. If you want to get a notification about new posts, I suggest you [follow me on dev.to](https://dev.to/mauricebrg) or add the [blog feed](https://aws-blog.de/index.xml) to your RSS reader.

&mdash; Maurice



