---
author: "Thomas Heinen"
title: "Hardware TOTP for AWS: Molto-2"
date: 2022-09-28
image: "img/2022/09/muhammad-zaqy-al-fattah-Lexcm-6FHRU-unsplash.png"
thumbnail: "img/2022/09/muhammad-zaqy-al-fattah-Lexcm-6FHRU-unsplash.png"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "security", "level-200", "well-architected"]
---
Everybody knows you should protect your AWS accounts (and other logins) with MFA against brute-force attacks. Most of the account providers use a standardized algorithm ([RFC 6238](https://www.rfc-editor.org/rfc/rfc6238)) to generate the famous six-digit TOTP codes for your login.

But where do you store those securely? Today, we will look at the alternatives and a specific device: The Molto-2.

<!--more-->

## Password Managers

Tools like [1Password](https://1password.com/), [LastPass](https://www.lastpass.com), or others offer to store your TOTP codes right with the primary login data. This way, you can easily log into your accounts and mitigate brute force attacks sufficiently.

But on the other hand, that means that everybody who gains access to your password manager or your computer can get around your security precautions. While having your machine or password safe compromised is scary enough, this way of keeping your MFA data safe can (and, in my opinion, should) violate internal compliance guidelines.

## AWS supplied tokens

As a way around this, AWS has been offering Gemalto tokens for a long time. For each of your AWS accounts, you can order one of these tokens and store them securely. If someone needs access, you can get that token (ideally, sign for it and have a four-eyes process in place) and use it.

![Gemalto 100](/img/2022/09/molto2-gemalto100.png)

But this poses some problems:

- you might have 100 accounts and do not want to store 100 tokens somewhere
- it costs a lot of money for multiple accounts
- Gemalto tokens are famous for being out of time sync when you need them
- you might want to secure not only the Root logins

## FIDO2 Keys

Apart from the six-digit codes you are probably used to, AWS also offers using devices implementing the cryptographic FIDO2 scheme for logging in. One of the most known vendors is [Yubico](https://www.yubico.com/), but as this is an open standard there are other vendors as well ([NitroKey](https://www.nitrokey.com/#comparison), 
[SoloKeys](https://solokeys.com/), ...)

This is basically a form of PKI, working with official certificates under the hood, making it the most secure way to protect your AWS accounts, but it comes with some limitations:

- FIDO2 keys do not work properly with the AWS CLI yet
- they are currently not as common as TOTP. AWS supports them, but other providers might not

I have been using a YubiKey for years now and would recommend it for protecting your personal AWS login, except for if you have policies which demand MFA approval (e.g. deletion of S3 objects) on the CLI.

## TOKEN2 "Molto 2"

I recently purchased an alternative, developed and built by Swiss company [TOKEN2](https://www.token2.com/home). They have a whole catalog of token generators and freely programmable single-purpose tokens.

But one device specifically caught my eye: the [Molto 2](https://www.token2.eu/shop/product/molto-2-v2-multi-profile-totp-programmable-hardware-token).

![TOKEN2 Molto 2](/img/2022/09/molto2.png)

This compact device has an eight-year battery life, can have its clock resynced, and stores up to 100 different TOTP profiles. It also acts as a USB keyboard, entering those codes at the press of a button.

The device ensures that you can only write/overwrite saved tokens and not read them, making attacks almost impossible. While the software to program new tokens is only available for Windows and is a bit retro, you only use it once per token.

## Use with AWS

As with your current TOTP solution, you start with your Root account or IAM user account configuration. But as the token does not have a camera to read the QR code, it gets a bit tricky. You need to download and unpack the software from the vendor first. Then, you can either take a picture of the QR code on the AWS page or enter the plain-text seed.

![Getting seed in AWS](/img/2022/09/molto2-aws-votp.png)

I prefer working with plain text, so I click on the __"Show secret key"__ link, and the random string is displayed. You can then copy this into the TOKEN2 programming software and choose a slot (0 - 99) plus a name for the profile. All other parameters can stay the same.

Then, you click on "Provision Profile," and your token will be on the small device.

![Configuring Molto 2](/img/2022/09/molto2-usbconfig.png)

If you are worried about what happens if the device breaks at some point, I suggest a second device or storing the secrets of your TOTP profiles in a safe place. 

That could be an [encrypted](https://www.kingston.com/de/usb-flash-drives/ironkey-kp200-encrypted-usb-flash-drive) [USB](https://www.kingston.com/en/usb-flash-drives/ironkey-d300s-encrypted-usb-flash-drive) [Stick](https://www.kingston.com/de/usb-flash-drives/ironkey-s1000-encrypted-usb-flash-drive) or simply a sealed envelope in your company vault.

## Summary

While I have worked with this device only for a short time, I already like it very much. Having someone compromise my laptop or mobile phone is one of my biggest worries, especially after recent waves of attacks even targeting well-known IT security people on YouTube.

While there is still the risk of someone eavesdropping during the short time of programming of the Molto 2, this reduces the risk enormously. If I made you curious, try this device or something similar (e.g., Reiner SCT tanJack deluxe, Reiner SCT Authenticator)

Pro:
- separate from your computer or mobile phone
- reasonable price, about 40 Euros
- 100 slots

Con:
- a bit thick for carrying around every day
- cumbersome programming of tokens
- no password lock - you get hold of it, you can use it

