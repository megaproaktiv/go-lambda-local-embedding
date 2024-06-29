---
title: "How do the AWS instances compare to commercially available graphic cards?"
author: "Roman Korneev, Philip Doerning"
date: 2022-02-04
toc: true
draft: false
image: "img/2022/02/BannerG5vs3090.jpg"
thumbnail: "img/2022/02/BannerG5vs3090.jpg"
categories: ["aws"]
tags: ["level-200", "g5.4xlarge", "RTX3090", "Unreal Engine", "NVIDIA A10G"]
summary: |
    In this post we would like to show a performance comparison between a g5-instance and the highend graphics card ASUS STRIX RTX3090 OC.
---

## Introduction
With the rise in popularity of blockchain technology leading to high demand for high-end-graphic cards and supply-chain shortages during the COVID-19 pandemic, the cost of acquiring a performant graphics card has been steadily increasing.[^1] Because the global chip shortage is expected to continue over the course of **2022**[^2] it is feasible to explore other ways to obtain access to high-end graphics cards instead of overpaying for a card or relying on an enthusiast friend to lend You one. The first thing that might come in mind is the use of disposable computational resources with the major Cloud providers. But how do the traditionally available high-end graphic cards compare to the cards offered in the cloud? In the following we will compare the newest **G-Family** instances available on **AWS**. Hereby, two publicly available **Unreal Engine benchmarking tools** will be run and evaluated on a **g5.4xlarge** instance. As a reference for the instance performance the enthusiast friend mentioned above has kindly provided his high-performance gaming set-up running a **GeForce RTX3090**.

## AWS set-up
AWS has recently made its new **EC2 G5** instances publicly available.[^3] The **G5** instances which are designed for graphic intensive workloads feature a **NVIDIA A10G Tensor Core GPU** and are powered through second gen. **AMD EPYC** processors.[^4] Being the obvious choice, the **G5** instances are expected to be at least as performant as a high-end gaming PC. Whether or not this statement checks out will be tested in the following. 
A **g5.4xlarge** instance utilizing **NVIDIA RTX Virtual Workstation – Windows Server 2019 AWS Marketplace AMI** was spun up on **AWS**.[^5] By using an **RDP** connection via **Microsoft Remote Desktop Application**, the connection to the instance was established. In the first step, the Windows Device Manager has been checked, to ensure that the **NVIDIA A10G** card is recognized by the OS on the instance. 

![DeviceManager](/img/2022/02/DeviceManager.png#center)

 _Figure 1: Screenshot of the Windows Device Manager running on the EC2 instance, showing the **NVIDIA A10G** card being utilized by OS_


As expected with the **NVIDIA AWS AMI** the **A10G** card is recognized correctly and has the necessary drivers installed. 
Two benchmarking tools based on **Unreal Engine** were installed to test the graphics performance of the **g5.4xlarge**  instance, the **Unreal Engine 4 Elemental DX12 Tech demo**[^6] and **Pure RayTracing Benchmark v1.51.**[^7]
Additionally, a technical demo utilizing **Unreal Engine 5** called **The Market of Light** was installed and benchmarked using the framerate capturing tool native to online gaming platform **Steam**. Because it is an interactive, game like demo, we want subjectively find out whether or not a **Unreal Engine 5** based game/demo is enjoyable being hosted on a VM in **AWS**.

## Results
With the chosen benchmarking tools utilizing different approaches to challenge the instance performance, the results of the tools will be discussed separately.
The reference high-end gaming PC utilizes an **ASUS Strix RTX3090**, an **Intel i7-7700k** processor, **GeForce RTX3090** and **16GB DDR4 RAM** at **2400 Mhz**. The benchmarks for each tool are measured on the **g5.4xlarge** instance and compared to the reference.

### Elemental DX12 Tech demo
The **Elemental DX12 Tech demo** is tested with no interaction between user and demo running a cinematic sequence rendered in real-time by the **Unreal Engine 4**. During the demo, the current **FPS** are shown along with latency. The demo was released in **Q3 2015**, so the used benchmark is quite old. Nevertheless, it still remains quite challenging on hardware. Remember, the **GeForce 900** series cards were introduced at a similar time, where the top-tier models remain desirable and cost even above the retail price stated in 2016 to this day.[^8]   

![Titan](/img/2022/02/GTX_titan.png)
_Figure 2: Picture of a used **NVIDIA GeForce GTX Titan X** sold in **Q1, 2022**_

The **Elemental DX12 Tech demo** has been run with **DirectX 12** in a resolution of **1600 x 900** pixels. 
With everything being said, the **g5.4xlarge**  instance is no slouch regarding performance during the cinematic demo, with an average performance of around **100 FPS** and **sub 7 ms** latency. The framerates are inconsistent and drop based on shown scenery. Fluctuations between **66 – 210 FPS** are present, but impressively the cinematic playback remains stable and smooth even looked at through an **RDP** connection. 

![elemental](/img/2022/02/elemental.png)

_Figure 3: A screenshot of **Elemental DX12 Tech demo** made on a **g5.4xlarge**  instance. The current **FPS** are shown on the right side, along with the latency in **[ms]**_

Compared to our high-end gaming PC reference the **g5.4xlarge**  instance is still underperforming. During the **Elemental DX12 Tech demo**, the reference reached stable **119 FPS** average, with excellent playback stability and occasional framerate fluctuations between **43 – 263 FPS**.

Concluding the first test run, we can confirm that the **g5.4xlarge**  instance with the **NVIDIA A10G** card is not quite as performing as our high-end reference, but delivers impressive performance nonetheless. A summary of all benchmark data is available at the end of the article.

### Pure RayTracing benchmark
Opposing to the **Elemental DX12 Tech demo**, the **Pure Raytracing benchmark** is a interactive benchmark, where a user can interact with the environment rendered by **Unreal Engine** in real time. For comparison reasons, we will stick to the cinematic demo performance also offered with the **Pure Raytracing benchmark**. With the benchmark being newer than the one in previous run, it potentially has more relevance for our use case. Some benchmark results are posted on the website already, but information about an **NVIDIA A10G** card is still missing.[^7] 
The benchmark was run three times on full HD resolution of **1920 x 1080** pixel, with varying quality settings between **usual**, **high** and **pro**.

![ray](/img/2022/02/ray.png)

_Figure 4: A screenshot of **Pure Raytracing benchmark** performed on a **g5.4xlarge** instance in interactive mode. Parameters, such as **resolution**, **quality settings**, **current** and **average FPS** are shown on the lower bottom of the figure_

Long story short, the **g5.4xlarge**  achieved quite good performance on the demo. The demo ran smooth in all cases, with excellent stability on usual setting but experiencing a few framerate drops on **high** and **pro** setting of around **20FPS**. This might also be caused by fluctuation in **RDP** connection performance.

Compared to our reference, the difference becomes more than clear. The reference PC reached on **average 1,4x times** better performance than our **g5.4xlarge**  instance. The playback stability and smoothness were unmatched, but also experiencing some framerate drops occasionally. 

Compared to some benchmark performances listed on[^7], the **GeForce A10G** performance is comparable to a **GeForce 2080 Ti Asus Rog Strix**. I provide you with the liberty to find out the price of that card for yourself at that point. Just a hint: at the current price in **Q1 2022** you could run a **g5.4xlarge** for quite a bit, without worrying of overshooting the price of the aforementioned **2080 Ti**. Not surprisingly, the performance of an overclocked **GeForce 3090** listed on[^7] was quite similar to our reference, being only **3 FPS** higher. The  **OC 3090** has been listed with a whopping **57 FPS** on **Pro** settings compared to our references **54 FPS**.

### The Market of Light
**The Market of Light demo** has been run with full HD resolution of **1920 x 1080** pixels. Surprisingly it was very much enjoyable to play on an instance hosted on **AWS** even by using a **RDP** connection. The **g5.4xlarge**  instance provided quite stable **32 FPS**, without major framerate oscillation, varying only between **16 – 33 FPS**. Well, at this point I should mention that a game being enjoyable is a subjective thing. The numbers provided by the **Steam** framerate capture tool are not looking bright in this case. Surely, our reference will be performing a bit better, but by how much? 
The reference put out **75 FPS** on average easily. But there is the thing: it suffered from quite severe FPS drops, with framerate fluctuating between **43 – 151 FPS**. This rendered the game as not quite enjoyable as is could be in the opinion of our enthusiast friend. Well, as mentioned above it is a subjective perception. 

![light](/img/2022/02/light.jpg)

_Figure 5: The Market of Light demo is shown. The demo could be played whilst being hosted on a **g5.4xlarge**  instance utilizing an RDP connection without major issues. Picture source[^9]_

### Comparison data
In the following, the benchmark data is summarized

**Reference RTX3090**

| Benchmark Name        | Avg. FPS | Low - High |    Resolution |   Stability |
| --------------------- | -------: | ---------: | ------------: | ----------: |
| `Elemental  demo`     |    `119` | `43 – 263` |  `1600 x 900` | `Excellent` |
| `Raytracing (Usual)`  |    `102` |            | `1920 x 1080` | `Excellent` |
| `Raytracing (High)`   |     `72` |            | `1920 x 1080` | `Excellent` |
| `Raytracing (Pro)`    |     `54` |            | `1920 x 1080` | `Excellent` |
| `The Market of Light` |     `75` | `43 – 151` | `1920 x 1080` |   `Average` |

 **g5.4xlarge**   
                                                           
| Benchmark Name        | Avg. FPS | Low - High |    Resolution |   Stability |
| --------------------- | -------: | ---------: | ------------: | ----------: |
| `Elemental  demo`     |    `100` | `66 – 210` |  `1600 x 900` | `Excellent` |
| `Raytracing (Usual)`  |     `71` |            | `1920 x 1080` | `Excellent` |
| `Raytracing (High)`   |     `48` |            | `1920 x 1080` | `Excellent` |
| `Raytracing (Pro)`    |     `38` |            | `1920 x 1080` | `Excellent` |
| `The Market of Light` |     `32` |  `16 – 33` | `1920 x 1080` | `Very Good` |


Concluding our little experiment, we can state the following based on the generated data:
- Not only it is possible to run **Unreal Engine 4** benchmarks and pixel streaming applications on **AWS**, it performs quite well
- Even the newest technologies (**Unreal Engine 5** being in early access build at time of writing in **Q1 2022**) can be run on a **g5.4xlarge**  instance already

A matter of cost should be mentioned here too: the **G5**-family is intended to be used for graphic intensive workloads, but more in a **Deep Learning** or **Machine Learning** environment. Therefore, the cost of running an instance of the **G5**-family could be quite expensive for an average gamer intending to experience the newest technologies like **Unreal Engine 5 first hand**.[^10] 

### Honorable mentions
For the chosen **Unreal Engine benchmarking tools** instances of the **AWS G4**-family were suggested. Feeling optimistic, we started-up an **g4dn.16xlarge** instance utilizing **NVIDIA RTX Virtual Workstation – Windows Server 2019 AWS Marketplace AMI** and installed the **Pure Raytracing benchmark** hoping to achieve comparable results to the **g5.4xlarge** instance. The results were sobering, with **g4dn.16xlarge** computational power not being enough to run the benchmark with acceptable framerate at **high** settings in interactive mode. 

![g4dn](/img/2022/02/g4dn.16xlarge.jpg)

_Figure 6: A screenshot of **Pure Raytracing benchmark** performed on a **g4dn.16xlarge** instance in interactive mode. With an **average** of only **17 FPS** the interactive mode is not as enjoyable as on a **g5.4xlarge** instance_

This furthermore illustrates the performance increase between the instance generations within an **AWS** instance-family and the high computational power required for the used benchmarking tools. 

## Sources

[^1]: https://wccftech.com/amd-radeon-NVIDIA-geforce-graphics-card-prices-reach-6-month-high-cost-up-to-83-over-msrp/

[^2]: https://www.digitaltrends.com/computing/NVIDIA-ceo-expects-chip-shortage-continue-throughout-2022/

[^3]: https://aws.amazon.com/about-aws/whats-new/2021/11/announcing-general-availability-amazon-ec2-g5-instances/?nc1=h_ls

[^4]: https://aws.amazon.com/ec2/instance-types/g5/

[^5]: https://aws.amazon.com/marketplace/pp/prodview-7qhjagotxzn22#pdp-pricing

[^6]: https://www.guru3d.com/files-details/unreal-engine-4-elemental-tech-demo-download.html

[^7]: https://marvizer.itch.io/pure-raytracing-benchmark

[^8]: https://en.wikipedia.org/wiki/GeForce_900_series

[^9]: https://store.steampowered.com/app/1691400/The_Market_of_Light/

[^10]: https://aws.amazon.com/ec2/pricing/?nc1=h_ls

