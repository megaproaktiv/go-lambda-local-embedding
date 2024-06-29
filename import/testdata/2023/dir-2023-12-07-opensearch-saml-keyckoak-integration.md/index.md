---
title: "Implementing SAML federation for Amazon OpenSearch Service with KeyCloak"
author: "Alexey Vidanov"
date: 2023-12-07
toc: true
draft: false
image: "img/2023/12/keycloak-opensearch.png"
thumbnail: "img/2023/12/keycloak-opensearch.png"
categories: ["aws"]
tags:
  [
    "aws", "opensearch", "level-400", "keycloak", "enterprise search", "SAML"
  ]
---

Welcome back to our series on implementing SAML Federation for Amazon OpenSearch Service. In [our previous post](https://www.tecracer.com/blog/2023/05/implementing-saml-federation-for-amazon-opensearch-service-with-onelogin..html), we explored setting up SAML Federation using OneLogin. Today, we'll focus on another popular identity provider - Keycloak.

Keycloak is an open-source Identity and Access Management solution, ideal for modern applications and services. We'll guide you through integrating Keycloak with Amazon OpenSearch Service to implement SAML Federation.

<!--more-->



If you're new to SAML Federation or Amazon OpenSearch Service, consider reading our previous post on [Implementing SAML Federation with OneLogin](https://www.tecracer.com/blog/2023/05/implementing-saml-federation-for-amazon-opensearch-service-with-onelogin..html) first. It covers the basics and provides a step-by-step setup guide.

## Prerequisites

Before we begin, there are a few prerequisites you need to have in place to follow along with this guide:

1. **Amazon OpenSearch Service Cluster**: You should have an Amazon OpenSearch Service cluster set up and running with Fine Grained Access Control enabled. If you're not sure how to do this, Amazon provides a [comprehensive guide](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/create-cluster.html) to get you started.
2. **Keycloak Instance**: You'll need a running instance of Keycloak. This could be a local development setup or a production instance, depending on your needs. If you don't have Keycloak set up yet, you can follow the [official Keycloak getting started guide](https://www.keycloak.org/getting-started). 
3. **Familiarity with SAML**: This guide assumes you have a basic understanding of SAML (Security Assertion Markup Language) and its role in authentication and authorization. If you're new to SAML, you might find our [previous post](https://www.tecracer.com/blog/2023/05/implementing-saml-federation-for-amazon-opensearch-service-with-onelogin..html) helpful. 
4. **Access to AWS Management Console**: You'll need access to the AWS Management Console with the necessary permissions to manage Amazon OpenSearch Service.

Once you have these prerequisites in place, you're ready to start implementing SAML Federation for Amazon OpenSearch Service with Keycloak!

## Keycloak-Specific Instructions

Let's start setting up SAML Federation with Keycloak

1. **Set Up a Keycloak Realm**: Create a new realm in Keycloak for your application.
   ![image-20231129190516970](/img/2023/12/image-20231129190516970.png)

2. **Create a Keycloak Client**: In your realm, create a new client for Amazon OpenSearch Service. Choose "SAML" as the client protocol.
   ![image-20231129190628712](/img/2023/12/image-20231129190628712.png)

    Make sure to select "SAML" as the client protocol. Use the endpoint URL for the Client ID.

   ![](/img/2023/12/image-20231129190957022.png)

   Set the "Sign Documents" options to "On" on the same screen in the "Signature and Encryption":
   ![image-20231129191523291](/img/2023/12/image-20231129191523291.png)

3. **Configure the Keycloak Client**: Adjust General Settings, turn on "Include AuthnStatement" and set "Sign Documents" to "On".
   ![image-20231129192029491](/img/2023/12/image-20231129192029491.png)

4. Configure client scopes. This will be used for the roles mapping from KeyCloak to Opensearch. By configuring the new mapper select there **Role List** 

   ![keycloak2](/img/2023/12/keycloak2.gif)

5. **Export Keycloak Metadata**: This will be used to configure SAML options in Amazon OpenSearch Service.

   ![image-20231129192546988](/img/2023/12/image-20231129192546988.png)

6. **Configure Amazon OpenSearch Service**: Log into the AWS Management Console and navigate to your OpenSearch Service domain. 

   

   ![](/img/2023/12/aws_console.png)  In the "Security" configuration, select "SAML" as the authentication type and upload the Keycloak metadata file you exported in the previous step.

   ![aws_console1](/img/2023/12/aws_console1.png)

7. **Set up roles and mapping users** 

   In the OpenSearch Dashboard, you can utilize existing roles or create new ones by navigating to the 'Security' section in the Management area.

   ![2023-12-07_14-32-13](/img/2023/12/2023-12-07_14-32-13.png)

   For learning purposes, we will add a user with the 'all_access' role in Keycloak. First, however, we need to map this role in the OpenSearch Dashboards. Go to 'Roles' and click on 'all_access'.

   ![image-20231207143413070](/img/2023/12/image-20231207143413070.png)

   Next, add the 'all_access' role to the backend roles mapping.

   ![image-20231207143702408](/img/2023/12/image-20231207143702408.png)

   Now, return to the Keycloak Realm and create a Realm Role named 'all_access'.  Finally, create an user for the OpenSearch Dashboards and assign the role.

   ![image-20231207165342853](/img/2023/12/image-20231207165342853.png)

   

8. **Test the Integration**: Log into OpenSearch Dashboards with Keycloak credentials to ensure proper setup.

Remember, these are high-level instructions and you may need to adjust them based on your specific Keycloak and AWS configurations. Always refer to the official Keycloak and AWS documentation for the most accurate and detailed information.

## Conclusion

Congratulations on successfully integrating Keycloak with Amazon OpenSearch Service for enhanced security and streamlined user management. This Single Sign-On (SSO) setup is extendable across AWS services.

For practical implementations, leverage Infrastructure as Code (IaC) tools like Terraform and AWS CloudFormation for efficient OpenSearch Service management. It's crucial to focus on security, availability, and cost efficiency.

Our tecRacer team offers expertise in deploying and managing OpenSearch Service infrastructure, ensuring robust security and scalability. We also provide integration support with various identity providers.

[Contact us](https://www.tecracer.com/en/consulting/amazon-opensearch-service/) for a consultation to optimize your Amazon OpenSearch Service deployment, ensuring a secure and efficient user experience.

**References**

For more information on the topics covered in this post, you may find the following resources helpful:

- [Keycloak Official Documentation](https://www.keycloak.org/documentation.html)
- [Amazon OpenSearch Service Documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/what-is.html)
- [SAML Federation with Amazon OpenSearch Service and OneLogin](https://www.tecracer.com/blog/2023/05/implementing-saml-federation-for-amazon-opensearch-service-with-onelogin.html)
- [Building SAML federation for Amazon OpenSearch Service with Okta](https://aws.amazon.com/de/blogs/architecture/building-saml-federation-for-amazon-opensearch-dashboards-with-okta)

-- [Alexey](https://www.linkedin.com/comm/mynetwork/discovery-see-all?usecase=PEOPLE_FOLLOWS&followMember=vidanov)
