---
title: "Enable Autocomplete for boto3 in VSCode"
author: "Maurice Borgmeier"
date: 2022-05-31
toc: false
draft: false
image: "img/2022/05/botostubs_2.png"
thumbnail: "img/2022/05/botostubs_2.png"
categories: ["aws"]
tags: ["level-200", "boto3", "vscode", "python"]

---

One of the less pleasant aspects of working with AWS using Python is the fact that most IDEs aren't able to natively support Autocomplete or IntelliSense for the AWS SDK for Python (boto3) because of the way boto3 is implemented. Today I'm going to show you how easy it has become to enable Autocomplete for boto3 in VSCode.

Before we come to the solution, let's talk about why native Autocomplete doesn't work with boto3. Usually, the Autocomplete functionality relies on performing static code analysis on the existing code and the installed dependencies. This means it analyzes the code without executing it. Here is the problem. If you look at the Python code installed when you run `pip install boto3`, you'll have trouble finding all the API methods you can use.

The reason is that the API methods and data structures are only shipped as JSON documents. Not all of them ship with boto3. You can find the majority in the underlying botocore library. The data directory in the boto3 package only contains the definitions for the resource API. Here's an example from the DynamoDB API in the botocore module.

![Example of a service description in the data directory](/img/2022/05/botostubs_3.png)

These documents are only parsed when boto3 or botocore are used, i.e., imported and instantiated somewhere. The SDK then constructs the methods and classes based on the JSON definition. From a technical perspective, this is pretty cool because you can update the SDK by changing the definition files. But it also has a few drawbacks. When you first instantiate the clients or resources, there is a performance penalty because the parsing needs to happen, and the data structures must be created. That's usually a one-time penalty because they're cached in memory. The more significant drawback is that Autocomplete can't discover these structures because they only exist in a *running* program - not something static code analysis can help with.

So... what are we going to do about this? Fortunately, some clever and dedicated people have created a solution that is now easy to use. The [boto3-stubs](https://pypi.org/project/boto3-stubs/) module ([Github](https://github.com/youtype/mypy_boto3_builder)) can be used to enable Autocomplete. The relatively new [VSCode extension for boto3](https://marketplace.visualstudio.com/items?itemName=Boto3typed.boto3-ide) makes using the library simpler than ever. Let's dive in. First, we'll create a working directory with a new virtual environment.

```shell
mkdir boto3-autocomplete && cd boto3-autocomplete
python3 -m venv .venv
# Activate the newly created environment
source .venv/bin/activate
pip install boto3
# Create an empty python script
touch aws_script.py
# Open the directory using vscode
code .
```

Now you should see VSCode in front of you, and we're going to start by adding some boilerplate code to the `aws_script.py`:

```python
import boto3

def dynamodb_fun():
    table = boto3.resource("dynamodb").Table("demotable")

    response = table.get_item(
        Key={
            "PK": "abc",
            "SK": "def"
        },
    )

def main():
    dynamodb_fun()

if __name__ == "__main__":
    main()
```

Next, it's time to install the boto3 VSCode Plugin, which you can do by navigating [here](https://marketplace.visualstudio.com/items?itemName=Boto3typed.boto3-ide) and hitting the "Install" button. It should take a few seconds, and the extension is installed. We need to open the command palette (Cmd + Shift + P) and start the Quickstart menu to complete the setup.

![Installation Step 1](/img/2022/05/botostubs_4.png)

A menu pops up, and in the bottom right corner, we'll want to click the "Install" button, which does all the necessary preparations.

![Installation Step 1](/img/2022/05/botostubs_5.png)

After a few seconds, a menu shows up to select which services we'll use. I've been a bit excessive and selected all of them, which means the setup takes a little bit longer.

![Installation Step 3](/img/2022/05/botostubs_6.png)

You'll see another menu with the progress bar and will have to wait a little bit for the tool to do its magic. In the background, it will install many packages, one for every service you selected. These packages contain "stubs" for the API methods. This means they install modules with all the classes that a "classic" implementation would have, just without the business logic. Here is an example - again from DynamoDB.

![Example stubs](/img/2022/05/botostubs_7.png)

As you can see here, the methods are empty but have the correct parameters and type hints to help with static code analysis. This is all we need to do. Now we can use the Autocomplete functionality as we're used to. It should usually pop out automatically; if that doesn't work - try Ctrl + Space, and you should see something like this. It provides data types and a link to the documentation as well.

![Autocomplete Request](/img/2022/05/botostubs_2.png)

What I like even more is that it also supports Autocompletion for the responses from the API. That's useful, and I don't need to refer to the docs quite as often. This fails where custom payloads are involved, such as in the `Item` property of the response. Under the hood, it uses typed dictionaries for this. These were [added to Python in version 3.8](https://peps.python.org/pep-0589/) - a handy feature that I had overlooked.

![Autocomplete Response](/img/2022/05/botostubs_1.png)

Let's see which dependencies the extension installed for us by running `pip freeze > requirements.txt`. Here's an excerpt:

```text
boto3==1.23.10
boto3-stubs==1.23.10
botocore==1.26.10
botocore-stubs==1.26.10
jmespath==1.0.0
mypy-boto3-accessanalyzer==1.23.0.post1
mypy-boto3-account==1.23.0.post1
mypy-boto3-acm==1.23.0.post1
...
```

My requirements.txt now has more than 300 lines, which seems a bit excessive, mainly because we don't need all these mypy-packages in production. If we just focus on Autocompletion, we can safely omit them. 

```shell
# Filter out mypy-boto3* and *stubs packages
pip freeze | grep -v "^mypy\-boto3" | grep -v "^.*stubs" > req
uirements.txt
```

These Mypy packages can do a lot more, though, we can also create type hints for our data structures with them, but that makes potentially bundling our packages for use in Lambda a bit more complex, and I haven't found a good solution for that yet, so we'll omit it for now. Maybe a tale for another time.

To summarize: I've shown you how you can enable Autocompletion for boto3 in VSCode using the correct extension and packages. I also explained why boto3 and botocore don't lend themselves to static code analysis out of the box and which steps the stubs library takes to mitigate this shortcoming during development.

Hopefully, you've learned something. For any questions, feedback, or concerns, feel free to reach out to me via the social media channels listed in my bio.

&mdash; Maurice