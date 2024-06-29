---
title: "Honeycode changed my mind about no-code"
author: "Maurice Borgmeier"
date: 2022-03-01
toc: false
draft: false
image: "img/2022/03/honeycode_logo.png"
thumbnail: "img/2022/03/honeycode_logo.png"
categories: ["aws"]
tags: ["level-200", "honeycode", "no-code"]
summary: |
    Most enterprises largely run on Excel. Imagine there was a tool that empowers spreadsheet specialists to build web and mobile apps without writing code. Amazon Honeycode tries to do that. We'll explore if it's as powerful as it sounds.
---

An inconvenient truth about most enterprises is that they largely run on Excel. It may be the single most critical software in many businesses. It's accessible, widely known, and can be *abused* to do almost anything. I've seen some impressive and scary excel sheets in the last few years, but I digress. Part of what makes it appealing is how easy it is to wrap your head around the spreadsheet format. What if you could use something similar to build mobile and web applications without writing code?

Amazon Honeycode proposes to do precisely that, and I think it has the potential to deliver on that promise. Before we continue, I should clarify that this is not an ad - there will be plenty of feedback to the team later on.

First, let's talk about the basics of the service as I understand them. When you launch it, you have to sign up using a separate account, but apparently, it also integrates with AWS SSO, which I didn't test, because it's only part of the paid plan. After signing up, a dashboard greets you that gives you an overview of your apps. Apps are not the primary entities that organize your resources, however. Apps are what your users will use.

The primary resource that you have to worry about is a workbook. For Excel users, this will be a familiar term. A workbook is a project folder that organizes the different components of Honeycode. Tables that look like spreadsheets but have some database-like capabilities hold your data. Apps are what your users use to interface with the data in the tables. You can use the Honeycode Builder to design screens that make up your app. Last but not least, there are automations. Changes to your data and the data itself can trigger these and perform many actions.

![Honecode Components](/img/2022/03/honeycode_components.jpeg)

You can create and edit your tables in a familiar spreadsheet-like interface, but these tables have some unique features that make them act more like databases. They can be linked to each other (think foreign-key relationships) and have data types. Tables also allow you to use many formulas you're familiar with from your favorite spreadsheet software.

![Honeycode Table](/img/2022/03/honeycode_table.png)

Now that we've set up our data model, we can build the user interface using the Honeycode builder. The builder allows you to create a UI from predefined components to link data and actions without writing code. In the top left corner, you can see that I can switch between the mobile and web view. By default, components are responsive and scale to your device's size, but you can also create different user interfaces for web and mobile. 

![Honeycode Builder](/img/2022/03/honeycode_builder.png)

Your finished app will also look more or less like this. You can share the app with your team when you've finished designing and tinkering with it. The screenshot below shows what the screen in the designer above looks like with my data. Don't blame the tool for my poor design skills - the app I built is just for me, and I wanted to get it done.

![Honeycode App](/img/2022/03/honeycode_app.png)

The last feature we should talk about here is automations. These are powerful and allow you to respond to changes in your data or based on your data. Automations always have a trigger, which can be one of these:

- Specific point in time
- A DateTime column in your data
- A row being added or removed in a table

Especially the second one is handy. You can run automations when you reach a specific timestamp in a DateTime column. I use the app from the screenshots to publish posts on _dev.to_ at a certain point in time. It may not sound like much, but it opens up many possibilities. That's because automations can have actions. You can, for example, send an email to some people, add rows to tables or call a custom webhook.

![Honecode Automation](/img/2022/03/honeycode_automation.png)

Webhooks are one way to make the no-code tool interoperate with more complex services that can implement stuff that is not possible in Honeycode, such as interfacing with third-party APIs, calling AWS services, and things like that. 

Now that we've gotten an impression of the service, I will talk about my experience with it. Note that at the time of writing this (February 2022), the service is still in Beta and has been since it was launched in June 2020, so I expect some rough edges. I like how easy it is to create tables, set data types, and link tables to each other. It feels immediately familiar, and they've done an excellent job at making it accessible.

The Honeycode builder is another big plus. It's straightforward to link your UI components to the data. It's almost tempting to say it's frictionless once you get the hang of it. I was especially impressed by how simple it was to create a form to enter data. You drag the form component on the canvas, select the table you want to insert a record in, the columns you want to fill, and it automatically generates the appropriate screen for you.

I also like automations because they let you trigger an automation based on the time in a column. It feels like I'm repeating myself, but that opens up many use cases. The same is true for the ability to call webhooks. You can put the more complex stuff in a Lambda function behind an API gateway and use a webhook to call that with data. That's very powerful.

There are, however, also a few things that I think could be improved:

- Date & Time formats
	- Support formats from countries other than the US. They do, in fact, exist (most of the world likes YYYY-MM-DD or DD-MM-YYYY)
	- Enable support for ISO8601 timestamps, including the time zone offset
- Charts
	- Part of what makes Excel valuable is how easy it is to create pretty charts. There is no support for that (yet)
	- Even basic support for bar-, line- and pie-charts would make significantly more use cases feasible
- [Pricing](https://www.honeycode.aws/pricing)
	- The basic tier is nice to have, but limiting workbooks to 2.500 rows per workbook seems extreme. A 10.000 row limit for the paid plus tier is also not that much.
	- Storage is cheap, and your customers most likely pay you plenty for their AWS account already
- Honeycode Builder
	- If you change the name of a screen to include characters it doesn't like, they're  ignored, there is no error
	- If you delete the home screen, there is no way to "promote" another screen. The whole thing bugs out
- Regions
	- Honeycode is only available in Oregon so far
	- For customers with data residency requirements, this is very limiting
- Velocity
	- As far as I can tell, the last [update announcement](https://honeycodecommunity.aws/c/announcements/7) is from January 2021
	- That's more than a year ago. Is this still in development?
	- Customer support in the Honeycode forums is still very responsive
- Localization / i18n
	- If you target an international user base, there should be some way to localize/translate the interface
	- Many people in the tech industry around the world understand English, but for businesses, it's a huge plus if the internal software is available in the local language
- Customization
	- It would be great if you could add images to the interface or embed third party content
- It's in Beta; I'm not sure if I can recommend it to customers

## Summary

I've built my first app with Honeycode and tried to explain the service and my thoughts on it. In summary, I think that Honeycode has much potential, and I'll try to keep up with it. I'd like to see the service grow. Hopefully, the development will pick up steam again soon, and the service leaves the Beta stage and launches in more regions.

Next week I'll also showcase the app I built with Honeycode. Stay tuned for that! If there are any questions, feedback, or concerns, feel free to reach out to me via the channels in my bio.

&mdash; Maurice

_My colleague Gernot also has [a few videos about Honeycode](https://www.youtube.com/playlist?list=PLFIhUqQSSW-byA6X1n29ItEPzXVtht8j5) on our Youtube Channel, feel free to check them out!_