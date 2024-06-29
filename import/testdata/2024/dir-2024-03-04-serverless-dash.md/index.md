---
title: "Deploying a Serverless Dash App with AWS SAM and Lambda"
author: "Maurice Borgmeier"
date: 2024-03-04
toc: false
draft: false
image: "img/2024/03/sam_dash_architecture.jpeg"
thumbnail: "img/2024/03/sam_dash_architecture.jpeg"
categories: ["aws"]
tags: ["level-400", "sam", "api-gateway", "dash", "lambda", "python"]
---

Today I'm going to show you how to deploy a Dash app in a Lambda Function behind an API Gateway. This setup is truly serverless and allows you to only pay for infrastructure when there is traffic, which is an ideal deployment model for small (internal) applications.

[Dash](https://dash.plotly.com/) is a Python framework that enables you to build interactive frontend applications without writing a single line of Javascript. Internally and in projects we like to use it in order to build a quick proof of concept for data driven applications because of the nice integration with [Plotly](https://plotly.com/examples/dashboards/) and [pandas](https://pandas.pydata.org/). For this post, I'm going to assume that you're already familiar with Dash and won't explain that part in detail. Instead, we'll focus on what's necessary to make it run serverless.

There are many options to deploy Serverless Applications in AWS and one of them is [SAM, the Serverless Application Model](https://aws.amazon.com/serverless/sam/). I chose to use it here, because it doesn't add too many layers of abstraction between what's being deployed and the code we write and our infrastructure is quite simple.

You can find the full application [on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sam-dash) if you want to deploy it yourself or just to follow along.

![Architecture Diagram](/img/2024/03/sam_dash_architecture.jpeg)

Let's talk about what we're going to deploy before we dive deeper. The architecture diagram is pretty simple, we're only using two AWS services: API Gateway as our entrypoint of the Frontend and Lambda to run the code in a Python 3.12 runtime.

Below, you can find a slightly abbreviated form of the SAM template that describes this infrastructure. As you can see it describes a pretty simple Lambda function that has two event sources. These refer to the root path and all other paths of the API gateway and basically route all requests to the API gateway to our Lambda function. Additionally we let the API Gateway know that we're handling some binary data trough the `BinaryMediaTypes` configuration. It's important that we increase both the Lambda function's memory as well as the timeout so that we get snappy responses and to account for potentially complex requests.

```yaml
# template.yml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Globals:
  Api:
    Function:
      Timeout: 30
      MemorySize: 1024
    BinaryMediaTypes:
      # The following is important for correct handling (b64)
      # of binary data, e.g. png, jpg
      - "*/*"

Resources:
  FrontendFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: frontend/
      Handler: app.lambda_handler
      Runtime: python3.12
      Architectures:
        - x86_64
      Events:
        # We want to capture both the root as well as
        # everything after the root path
        RootIntegration:
          Type: Api
          Properties:
            Path: "/"
            Method: ANY
        ProxyIntegration:
          Type: Api
          Properties:
            Path: "/{proxy+}"
            Method: ANY
# ...
```

Since the AWS part of the infrastructure is rather boring, let's talk about what's necessary to make Dash play nicely with the API Gateway. Both are originally not designed to work with each other - Dash is intended to be executed as a long running process in a webserver and the API gateway is not _really_ intended to serve this kind of static and dynamic website content.

When the API Gateway calls the Lambda function, it embeds the information from the request in a JSON object that doesn't really match what Dash or its preferred webserver, Flask, expect. Dash would like to see a request matching the [web server gateway interface (WSGI)](https://en.wikipedia.org/wiki/Web_Server_Gateway_Interface) specification. Fortunately there's a nice community project called [apig-wsgi](https://pypi.org/project/apig-wsgi/) that aims to bridge that gap and convert one into the other.

This already does 95% of what we need, we just have to adjust it a little bit to handle some corner cases. Before we do that, let's take a brief look at the Dash app that I defined:

```python
# frontend/dash_app.py
from dash import Dash, html, dcc, Input, Output, callback

def build_app(dash_kwargs: dict = None) -> Dash:

    dash_kwargs = dash_kwargs or {}

    app = Dash(
        name=__name__,
        **dash_kwargs,
    )

    app.layout = html.Div(
        children=[
            html.H1(children="Hello World!"),
            html.P("This is a dash app running on a serverless backend."),
            html.Img(src="assets/architecture.jpeg", width="500px"),
            # ...
        ],
        # ....
    )

    return app
```

The `build_app` function creates a Dash app that displays some text, an image, and some dynamic functionality. I've omitted some of the code for brevity, you can find the full code [on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sam-dash). I wrapped the instantiation in a function that allows me to pass additional arguments to the apps' constructor.

Before we talk about the actual Lambda handler that does the conversion and invocation of that app, I want to introduce you to a few quirks of the API Gateway event. You may already be familiar with the concept of a stage in the API Gateway. When you deploy an API, you always deploy it to a stage which is added as a suffix to the URL:

```
https://<id>.execute-api.<region>.amazonaws.com/<stage>/
```

When the Lambda function is called, API Gateway omits the `/<stage>/` part of the path in the top level `path` attribute of the event, which is a problem if you tell Dash to use that as part of its internal URL management logic, because any link it generates needs the `/<stage>/` prefix otherwise they won't work. Fortunately the _actual_ path is also available as part of the `requestContext`.

This gets a bit more confusing if you're using a custom domain for the API Gateway and there suddenly is no prefix anymore. If your request is coming from that kind of source, we shouldn't add the `/<stage>/` prefix to URLs that we return. Long story short, we want to replace the `path` attribute at the top level with the _original_ path and have a separate handling for URLs that we send to the outside world. Here's some pseudocode of what we need to do:

```python
if original_path.startswith("/<stage>/"):
    return build_dash_app_that_adds_stage_prefix()
else:
    return build_dash_app_without_stage_prefix()
```

The actual implementation is a bit more verbose than that, but it gets the job done.

```python
# frontend/app.py
import json

from functools import lru_cache
from apig_wsgi import make_lambda_handler
from dash_app import build_app


@lru_cache(maxsize=5)
def build_handler(url_prefix: str) -> "Dash":

    # If there's no prefix, it's a custom domain
    if url_prefix is None or url_prefix == "":
        return make_lambda_handler(wsgi_app=build_app().server, binary_support=True)

    # If there's a prefix we're dealing with an API gateway stage
    # and need to return the appropriate urls.
    return make_lambda_handler(
        wsgi_app=build_app({"url_base_pathname": url_prefix}).server,
        binary_support=True,
    )


def get_raw_path(apigw_event: dict) -> str:
	# ...
    return apigw_event.get("requestContext", {}).get("path", apigw_event["path"])


def get_url_prefix(apigw_event: dict) -> str:
	# ...

    apigw_stage_name = apigw_event["requestContext"]["stage"]
    prefix = f"/{apigw_stage_name}/"
    raw_path = get_raw_path(apigw_event)

    if raw_path.startswith(prefix):
        return prefix

    return ""


def lambda_handler(
    event: dict[str, "Any"], context: dict[str, "Any"]
) -> dict[str, "Any"]:

    # We need the path with the stage prefix, which the API gateway hides a bit.
    event["path"] = get_raw_path(event)
    handle_event = build_handler(get_url_prefix(event))

    response = handle_event(event, context)
    return response

```

This basically ensures that our frontend works with plain API gateway URLs as well as custom domains. If you're only using custom domains without a prefix, you can replace all of that with the following. Not everyone wants to deal with or has custom domains though.

```python
lambda_handler = make_lambda_handler(
	wsgi_app=build_app().server,
	binary_support=True
)
```

Now that we've walked through most of the infrastructure code, let's deploy our app:

```terminal
$ sam build
$ sam deploy
```

In the output you'll find the URL of the API Gateway, just click on it and you should be greeted with this website:

![Webapp Screenshot](/img/2024/03/sam_dash_screenshot.png)

As you can see, this setup allows us to service some static content in the form of text and images, but also some interactive functionality with the date picker that updates the text below it through this simple Python function:

```python
@callback(
    Output("birthdate-output", "children"),
    Input("birthdate-picker", "date"),
    prevent_initial_call=True,
)
def calculate_days_since_birth(birthday_iso_string):

    now = datetime.now().date()
    birthday = date.fromisoformat(birthday_iso_string)

    age_in_days = (now - birthday).days
    output = ["You're ", html.Strong(age_in_days), " days young!"]
    return output
```

While this setup is cool, it has some limitations:

- The API Gateway only listens to HTTPS traffic, that means there's no easy way to add an HTTP to HTTPS redirect (but you could add CloudFront)
- There is a response size limit of 6 MB, that means it's not suitable to handle large objects
- API Gateway enforces a timeout of 29 seconds, so longer-running computations aren't possible
- The Lambda package size limits how many libraries you can use and in this small example I didn't bundle pandas for that reason. Using layers or Image-based Lambdas allows you to circumvent this limit.

There are some things that can be done to make this more scalable, e.g. serving static assets from CloudFront or S3 or adding Cognito Authentication and I may get into that in a future blog post.

For now that's it. I hope you learned something new today and don't forget to check out the code [on Github](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sam-dash).

&mdash; Maurice


---

Other articles in this series:

- Deploying a Serverless Dash App with AWS SAM and Lambda
- [Adding Basic Authentication to the Serverless Dash App](https://www.tecracer.com/blog/2024/03/adding-basic-authentication-to-the-serverless-dash-app.html)
- [Build a Serverless S3 Explorer with Dash](https://www.tecracer.com/blog/2024/04/build-a-serverless-s3-explorer-with-dash.html)