---
title: "Adding Basic Authentication to the Serverless Dash App "
author: "Maurice Borgmeier"
date: 2024-03-20
toc: false
draft: false
image: "img/2024/03/kutan-ural-MZPwImQUDM0-unsplash.jpg"
thumbnail: "img/2024/03/kutan-ural-MZPwImQUDM0-unsplash.jpg"
categories: ["aws"]
tags: ["level-400", "sam", "api-gateway", "dash", "lambda", "python"]
summary: I'll teach you how to add interactive basic auth to the Serverless Dash app that we deployed recently.
---
 
In a recent [post](https://www.tecracer.com/blog/2024/03/deploying-a-serverless-dash-app-with-aws-sam-and-lambda.html) explained how to set up a Dash app in a completely serverless manner. If you haven't read that one yet, you should - otherwise this won't make much sense because we'll continue where we left off.
 
At the end of that post we successfully hosted a Dash App in a Lambda function behind an API Gateway. That's already a good basis but unless you want your webapp to be open to the public, you probably want to add some form of authentication. Today I'm going to explain how to add HTTP Basic Auth, in a future post I'll dive into Cognito. 
 
HTTP Basic Auth has been around for a very long time ([RFC 7235](https://datatracker.ietf.org/doc/html/rfc7235)). The general idea is pretty simple. If a website requires credentials, you send them in the authorization header. The authorization string starts with *Basic* and is followed by the base64 encoded version of concatenating username and password with a colon as the separator. It looks something like this. 
 
```text 
Authorization: Basic base64(<username>:<password>) 
``` 
 
Obviously this should be only used when communicating over an encrypted HTTPS connection, otherwise the credentials are sent in clear text over the internet. Given that the API Gateway supports exclusively HTTPS, we can take that for granted.

But how does a user enter the credentials when accessing the website? 
The first option is through the url, like this: 
 
```text 
https://<username>:<password>@website.com/
``` 
 
The browser will then automatically encode this correctly as basic auth information. But that's not very user friendly and potentially logs the credentials in the browser history so it should be avoided. The other option is that the server asks for authentication by sending a response with the HTTP status code 401 and this header:
 
```text 
WWW-Authenticate: Basic realm="mywebsite" 
``` 

If the server sends this header, the browser will prompt for the credentials. The `realm=...` part is optional and only interpreted by some browsers. A browser that supports it will display the realm info as part of the login prompt. This Flow-Diagram from the [mdn web docs about HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) visualizes this quite well.

![MDN Basic Auth Flow Diagram](/img/2024/03/mdn-http-auth-sequence-diagram.png)

 
Here's what the login prompt looks like:

![Screenshot: Login Prompt](/img/2024/03/basic_auth_login_screenshot.png)

Granted, this login window isn't very pretty but it gets the job done and protects our website from unauthorized access.
 
Now that we understand this flow, let's start building it. If you're already familiar with Dash, you may know that it has a [basic auth extension](https://pypi.org/project/dash-auth/) that seems like it would be useful here. Unfortunately it isn't. The API Gateway will [rename the `WWWW-Authenticate` header](https://stackoverflow.com/questions/58037317/) if it's returned from the backend to resolve [potential ambiguity](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-known-issues.html#api-gateway-known-issues-rest-apis), which means the browser won't prompt for credentials. This means the straightforward solution won't work. 
 
Instead, we have to rely on a custom authorizer and a gateway response. This approach is inspired by a [medium article](https://medium.com/@Da_vidgf/http-basic-auth-with-api-gateway-and-serverless-5ae14ad0a270) that I adapted and extended. The custom authorizer is a Lambda function that we'll write to decode the Authorization header and check the credentials. It will support multiple credential backends: hardcoded credentials, the SSM Parameter Store and a Secretsmanager Secret. But let's not get ahead of ourselves. 
 
![Architecture](/img/2024/03/sam_basicauth_architecture.png)

In addition to the Lambda function, we need to configure a [gateway response](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-gatewayResponse-definition.html). These gateway responses allow us to send a response with the `WWW-Authenticate` header that we need for the prompt. Gateway responses are employed if the backend or an authorizer don't send a response. It uses a simple pattern matching based on the HTTP status code.
 
To implement our solution, we add a few more resources to our SAM app. The authorizer function, an SSM parameter for the credentials, and the gateway response. 

```yaml 
# template.yaml
Resources:
  #...
  AuthorizerFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: basic_auth_authorizer/
      Handler: basic_auth.lambda_handler
      Environment:
        Variables:
          CREDENTIAL_PROVIDER_NAME: SSM
          SSM_CREDENTIAL_PARAMETER_NAME: !Ref SsmParameterWithCredentials
          # ...
      Policies:
        - SSMParameterReadPolicy:
            ParameterName: !Ref SsmParameterWithCredentials
	  # ...
  SsmParameterWithCredentials:
    Type: AWS::SSM::Parameter
    Properties:
      Type: String
      # JSON-Object, e.g. {"username": "password"}
      Value: "{}"
  BasicAuthPrompt:
    Type: AWS::ApiGateway::GatewayResponse
    Properties:
      ResponseType: UNAUTHORIZED
      RestApiId: !Ref ServerlessRestApi
      StatusCode: "401"
      ResponseParameters:
        gatewayresponse.header.WWW-Authenticate: !Sub '''Basic realm="${AuthenticationPrompt}"'''
``` 
 
The code for this setup is [available on Github again](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sam-dash-basicauth). I'm not going to walk through all of the authorizer code here, feel free to check out the full [implementation](https://github.com/MauriceBrg/aws-blog.de-projects/blob/master/sam-dash-basicauth/basic_auth_authorizer/basic_auth.py), it should be relatively easy to read.
 
```python 
# basic_auth_authorizer/basic_auth.py
def lambda_handler(event, _context):

    credential_provider = CREDENTIAL_PROVIDER_NAME_TO_CREDENTIAL_PROVIDER.get(
        os.environ.get(ENV_CREDENTIAL_PROVIDER_NAME, "HARDCODED")
    )

    valid_credentials = credential_provider()

    try:
        username, password = get_username_and_password_from_header(event)

        correct_password = valid_credentials.get(username)

        if password == correct_password:

            prefix, stage, *_ = event["methodArn"].split("/")

            all_resources_arn = f"{prefix}/{stage}/*"

            policy = {
                "principalId": username,
                "policyDocument": {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Action": "execute-api:Invoke",
                            "Effect": "Allow",
                            "Resource": all_resources_arn,
                        }
                    ],
                },
            }
            return policy
        raise UnauthorizedException()

    except Exception:
        return "Unauthorized"
``` 
 
As you can see, the `lambda_handler` function first selects the credential provider based on an environment variable and looks up the supported credentials. Next, it extracts the authorization string from the event and parses it into the supplied username and password. If anything goes wrong here or the header is missing, an exception is raised and the function returns *Unauthorized*. This will then trigger the login prompt through the gateway response.

If we were able to fetch valid credentials and the supplied credentials match, we create an IAM policy that grants access to the API gateway. This policy document is then returned and the API gateway caches the response for a period of time (5 minutes by default). The policy I implemented here is very broad and grants full read/write access to that stage of the API gateway. If you need a more granular policy, you can achieve that through a more detailed resource section.
 
I chose to make the credential provider plugable and currently the code supports hard-coded credentials, an SSM parameter or a SecretsManager secret. More details on their respective configurations can be found in the [Github repo](https://github.com/MauriceBrg/aws-blog.de-projects/tree/master/sam-dash-basicauth#authentication-backends).

```python
# basic_auth_authorizer/basic_auth.py
CREDENTIAL_PROVIDER_NAME_TO_CREDENTIAL_PROVIDER: dict[
    str, Callable[[], dict[str, str]]
] = {
    # "HARDCODED": lambda: {"avoid": "using_me"},  # You really shouldn't use these.
    "SSM": get_credentials_from_ssm_parameter,
    "SECRETS_MANAGER": get_credentials_from_secrets_manager,
}
```
 
As the last step, we just need to tell our API gateway to use the new authorizer function, which I'm doing in the `Global` section of the SAM template. 
 
 ```yaml 
# template.yaml
Globals:
  # ...
  Api:
    Auth:
      Authorizers:
        BasicAuth:
          FunctionArn: !GetAtt AuthorizerFunction.Arn
      DefaultAuthorizer: BasicAuth
``` 
 
After running `sam build && sam deploy` we can test it. When you access the website you should be prompted to login and after to enter the credentials you'll be allowed to see the webapp.
 
This setup ensures that only authenticated access to the webapp is possible. There are however a few drawbacks. Username and passwords are stored in plain text (albeit encrypted at rest depending on the backend), which is not ideal for a production system. It would be safer to salt and hash the saved password in order to store a derivative of the original to check against without storing the password itself.
 
This could be implemented, but there's a much better solution if you need something more serious. It's possible to integrate the serverless Dash app with Cognito and I'll show you how to do that in the future. The other problem is the lack of proper session management. The app only knows that someone is logged in and not many more details, which makes authorization difficult.
 
Now that I've outlined the weaknesses that you should be aware of, let's talk about what this is intended to be used for. If you want to build a quick PoC or just avoid exposing the app to unauthenticated users, this setup is a good start. It has a limited complexity, is easy to integrate and in addition to that not really Dash-specific, i.e. we didn't touch the Dash app once. Sure it's not a fancy solution, but that's not always necessary.

Hopefully you learned something from this post, stay tuned for the implementation with Cognito based authentication.
 
&mdash; Maurice

---

Title Photo by <a href="https://unsplash.com/@kutanural">Kutan Ural</a> on <a href="https://unsplash.com/photos/royal-guard-guarding-the-buckingham-palace-MZPwImQUDM0">Unsplash</a>

---

Other articles in this series

- [Deploying a Serverless Dash App with AWS SAM and Lambda](https://www.tecracer.com/blog/2024/03/deploying-a-serverless-dash-app-with-aws-sam-and-lambda.html)
- Adding Basic Authentication to the Serverless Dash App
- [Build a Serverless S3 Explorer with Dash](https://www.tecracer.com/blog/2024/04/build-a-serverless-s3-explorer-with-dash.html)