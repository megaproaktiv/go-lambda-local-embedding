---
author: "João Azevedo"
title: "First Steps with Amazon Braket SDK"
date: 2022-10-10
image: "img/2022/10/arrabida_sunrise_banner.png"
thumbnail: "img/2022/10/arrabida_sunrise_banner.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "braket", "quantum computing", "qubit", "sdk", "tutorial", "bell state"]
---
*In this article, we will install and deploy a circuit with a Bell State using AWS Braket SDK.*
*Some minimal knowledge requirements are expected from the reader, namely concepts such as: Qubit, Superposition, Quantum Logical Gates (and a bit of math by association), Bell State, Python and perhaps AWS CLI.*

<!--more-->

## Preface 
**Who is this article directed to?** 
This article is directed at people who are familiar with [QC](https://en.wikipedia.org/wiki/Quantum_computing) and perhaps already know a bit how to use the AWS console. It will give a glimpse on how to navigate around Amazon Braket. It will also address how to install and use Amazon Braket SDK (CLI), followed by the implementation of a small circuit example of a [Bell state](https://en.wikipedia.org/wiki/Bell_state).

**What this article is not about?**
The purpose of this article **is not to explain what quantum computing** is. 


## Let's Start Off 

The installation and usage of Amazon Braket SDK takes place in the big stage that is [AWS CLI (Command Line Interface)](https://aws.amazon.com/cli/).
### What is AWS CLI? 
It is a tool that allows us to manage AWS sevices through the command line and, for instance, to run automation scripts.
It is expected that you already have the AWS CLI in place before we jump on to the next section. Installation guidelines can be found [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).
Furthermore, the SDK allows also to run tasks on a local simulator - a simulator that is ran in your local enviroment.

## Tool Requirements
First off, we will need to take care of some dependencies, each of which are assumed to be done before the SDK installation. Please check the following hyperlinks as per requirement:
- [Python 3.7.2 or greater](https://www.python.org/downloads/)
- [Git tools](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [IAM User or a role with the required permissions to access Braket](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)
- [Boto3 and setting up AWS Credentials](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/quickstart.html)


## What is Amazon Braket?
Amazon Braket is a service provided by AWS which allows users to run jobs and tasks of quantum algorithms in physical devices and simulators. These algorithms are expressed using python to write out how to build circuits with gates – such as a solution for the [Elitzer-Vaidman Problem](https://en.wikipedia.org/wiki/Elitzur–Vaidman_bomb_tester) – and techniques – such as depth reduction, via transpilation. We should also take into account each circuit’s number of qubits, their layout, noise and type of employed technology. 

>*Why is this important?* 
>Understanding these parameters allows for the writing of a more suitable program, rendering it an effective run. In practice, this means cost reduction, time reduction and noise reduction. While it is important to know this, it won't have much relevance for our simple circuit later.


## Braket Structure
As mentioned, we can deploy our programs in either simulators or QPUs (Quantum Processing Units), each of which are represented by their own ARN. These physical devices are hosted by other companies with which AWS operates.

> **ARN?**
>An Amazon Resource Name (or ARN) is a resource's identification path that allows us to reference it across all AWS.

![AWSQPUs](/img/2022/10/qpus.png)
Figure 1 – Screenshot of available QPUs and Simulators taken from the AWS Console.


Amazon Braket is part of the [AWS Free tier](https://aws.amazon.com/braket/pricing/) catalog which in practice means that we have 1 hour per month to run any job without incurring costs as long as we use AWS Simulators (SV1, DM1, TN1) or some combinations of these.

<img src="/img/2022/10/qpu_example.jpg" alt="QPU_Example" width="300"/>

Figure 2 - Example of a QPU: [D-Wave's 2000Q](https://www.dwavesys.com/company/newsroom/press-release/d-wave-makes-new-lower-noise-quantum-processor-available-in-leap/) - one of the quantum computers used by AWS Braket - without the whole cryogenic apparatus. 

## Installing the SDK
Let's open our **command line** – you should have it already opened if you followed the Requirements section – and execute the following
```console
$pip install amazon-braket-sdk
```

And, if we ever need to update the SDK, just execute the following:
```console
$pip install amazon-braket-sdk --upgrade --upgrade-strategy eager
```


## Using the SDK
### Call your Profile
Through the command line you should add, if you haven’t already, an adequate profile, by means of changing the *credentials* file:

```console
$nano ~/.aws/credentials
```

Which should have the following layout:

![Credentials](/img/2022/10/credentials.png)
Figure 3 - Credentials file change using the command line.


>**aws_credentials_id? aws_secrete_access_key?**
>On your console you can find or create new credentials by...
> - Going to the [console](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/learn-whats-new.html) and click on the menu with your username in the upper right corner;
> - Press *Security Credentials*;
> - Under *Access keys for CLI, SDK, & API access*, click on *Create access key*.

We can go ahead use the *us-east-1* region. 

Afterwards, configure a named profile. In this example the profile is named *joao*, thus:

```console
$aws configure --profile joao
```

### Region
When choosing a [QPU](https://en.wikipedia.org/wiki/Quantum_computing), it is rather important to beware picking an adequate region, according to the table below. 

![Regions and QPUs](/img/2022/10/region_qpu.png)
Figure 4 - Region per QPU taken from the Console.

Furthermore, simulators are available in the following regions:
- SV1 (State Vector Simulator) - eu-west-2, us-east-1, us-west-1, us-west-2;
- TN1 (Tensor Network Simulator) - eu-west-2, us-east-1, us-west-2;
- DM1 (Density Matrix Simulator) - eu-west-2, us-east-1, us-west-1, us-west-2;

More information can be seen at the AWS console by choosing the Amazon Braket service and then selecting *Devices*.

## Bell State Design

Notwithstanding the sections above, we now have all the necessary tools an knowledge to start writing code. Let us then design a standard circuit of two qubits: we will take **two ground states** q<sub>1</sub> and q<sub>2</sub>, represented by |q<sub>2</sub> q<sub>1</sub>⟩.

First, we apply an Hadamard gate to q<sub>1</sub> which will transform the state into a superposition:


<img src="/img/2022/10/haddamard.png" alt="Haddamard" width="300"/>


Which, in relation to the second qubit, q<sub>2</sub>, we get that:

<img src="/img/2022/10/2qubits.png" alt="2Qubits" width="200"/>


In turn, we then apply CNOT gate with q<sub>2</sub> as a target and q<sub>1</sub> as a control, effectively converting our initial couple of qubits into one Bell state (out of four…):

<img src="/img/2022/10/cnot.png" alt="CNOT" width="400"/>


By result of these operations is a circuit as follows:

![Bell Circuit](/img/2022/10/bell_circtuit.png)
Figure 5 - Circuit encoding one Bell State. 

### Code 

*What device shall we use?*
Let’s go with the simplest one, as per figure 1, the state vector simulator does the trick. 
To use it we need to fetch our intended device ARN, which can be easily done by searching for the devices on the Braket Service through the [Console](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/learn-whats-new.html). 

There are many ways to deploy the code, for instance, we can either write on some code editor such as VSCode and run it there, or we could write and run it on Jupyter Notebook, or even run it through the command line. Here, we will use the latter.

Now, create a file named *example.py* and write the following:

```python
import braket._sdk as braket_sdk
import boto3
from braket.aws import AwsDevice
from braket.circuits import Circuit
import numpy as np
import matplotlib.pyplot as plt

sv_device = AwsDevice("arn:aws:braket:::device/quantum-simulator/amazon/sv1")

bell = Circuit().h(0).cnot(control=0,target=1)
```

The device ARN, as we wrote for the *sv_device*, can be obtained by going back to the console and running:
```console
$aws braket search-devices --filters
```

<img src="/img/2022/10/devices_arn.png" alt="Devices ARN" width="500"/>

Figure 6 - Output of a list of devices available in our selected region (*us-east-1*), as a result of the previous command.

We have also now defined a variable called *bell* where we instantiated a new circuit object, *Circuit()*, and to which we applied the procedure described in figure 5.

```python
task = sv_device.run(bell, shots=1000)
```
A shot is a single execution of the quantum algorithm. Therefore, the higher the number of shots the higher the accuracy.  

```python
print (bell) 
print (task.result().measurement_counts)
```
The first print will output a ASCII visual of our circuit in the command line. And, the latter, will print out the result of 1000 shots for the entangled states.

```python
plt.bar(counts.keys(), counts.values())
plt.xlabel('bitstrings')
plt.ylabel('counts')
plt.show()
```
This last piece will gives a prettier visual of the results. 

### Results 
We can now save the changes made to the file and run it at the command line:

>Reminder: be sure to be in the same directoy as the the file you are trying to access.

```console
$ python example.py
```

Finally, we get our long-waited results – not that long though because it only took a few seconds. 
```console
$ Counter({'11':508, '00':492})
```
The line above - the output in the command line - shows the number of counts per entangled pair of qubits.
And below we have a column chart representation of the same thing.

<img src="/img/2022/10/counts.png" alt="Count Chart" width="500"/>

Figure 7 – Number of counts per entagled pair.

As we can see, from the |00⟩ state we got a Bell state |Φ<sup>+</sup>⟩, where we have nearly 50% probability of being measured in the state |00⟩ and 50% of being measured in |11⟩. 
We get a state truly entangled which in practice means that measuring one qubit will determine the state of the other qubit: for instance, if we measure the q<sub>1</sub> qubit and we obtain a |0⟩, then the other entangled qubit must also be **necessarily** a |0⟩.


## Final Remarks 
Congratulations! 🎉 You’ve created and run your first quantum circuit through Amazon Braket SDK! Hopefully this showed you how easy it is to deploy a quantum circuit and to obtain results straight from the command line.

tecRacer has a Quantum Computing team! Apart from architecting solutions, we also give workshops as well on the topic. For more information feel free to visit us at [our website](https://www.tecracer.com/training/workshop-quantencomputing/) or contact us 😁 via <quantumcomputing@tecracer.de> 

