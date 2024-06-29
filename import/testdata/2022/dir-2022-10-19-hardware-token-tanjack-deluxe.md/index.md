---
author: "Thomas Heinen"
title: "Hardware TOTP for AWS: Reiner SCT tanJack Deluxe"
date: 2022-10-19
image: "img/2022/10/markus-krisetya-Vkp9wg-VAsQ-unsplash.jpg"
thumbnail: "img/2022/10/markus-krisetya-Vkp9wg-VAsQ-unsplash.jpg"
toc: false
draft: false
categories: ["aws"]
tags: ["aws", "security", "level-200", "well-architected"]
---
Even when safely storing your MFA tokens using the [Token2 Molto-2](https://aws-blog.com/2022/09/hardware-totp-for-aws-molto-2.html) device, some things are not quite optimal. You have to use special Windows-only software to program new accounts, it is not PIN-protected, and things could be better in terms of usability.

If you have a bit more of a budget, the [Reiner SCT tanJack Deluxe](https://shop.reiner-sct.com/tan-generatoren-fuer-sicheres-online-banking/tanjack-deluxe?locale=en) might solve your problems. Let's have a look at this device.

<!--more-->

## Usability

The tanJack Deluxe offers a streamlined experience for managing your MFA tokens. To add a new one, use the integrated camera to take a picture of the QR code, just like with your phone. You will then be able to change the title and user name displayed on the device, and you are ready to go.

All this is made extremely easy by the device's touch screen, allowing easy management. You can also pin often-used accounts as favorite to quickly access them. While there is no official statement on the maximum capacity for MFA tokens, the vendor lists it as "100+", which puts it on par with the Molto 2.

![tanJack Deluxe (image by Reiner SCT)](/img/2022/10/tanjack-deluxe.png)

A device like this won't have the extensive battery life of a minimalist device as the competition offers, but Reiner SCT thought ahead and bundles a USB-C connected charging cradle to make this easier.

## Security

Other devices openly display current MFA codes on their display, meaning if someone gets access to the device, they can just use them.

With the tanJack Deluxe, you can set a PIN for accessing the TOTP functionality. And, as this is a secure device, entering the wrong PIN five times in a row will automatically wipe the device.

This additional protection makes it highly secure, especially as all USB connectivity only uses the charging pins, and there is no data connectivity at all.

But if you ever used an embedded device, this will immediately result in one question: How would you update the device firmware if there are any bugs? Reiner SCT uses the same mode as with other QR-capable readers: they offer a "QR Code Movie" for their devices. You can enter the device firmware update mode and let it ingest new code using the camera. 

Currently there is no firmware update available, but I expect Reiner SCT as a security-conscious company to use some code signing to avoid attacks where manipulated firmware is uploaded.

## Addon Benefits

One main criticism about the usual Gemalto devices is the problem of time drift. If the onboard RTC clocks get sufficiently out of sync with the remote website, its codes will not work anymore.

For this case, Reiner SCT offers a [specific web page](http://www.reiner-sct.com/sync) where the device time can be updated. Again, a QR code is recorded with the camera to resynchronize time. An critical security feature is that the QR code does not simply include the time because this would allow getting codes from the past/future using a manipulated code.

The main purpose of the device is not generating TOTP MFA codes at all - it is primarily a device for secure banking. As such, you can insert your bank card into its card reader slot and generate TANs for online banking (ChipTAN QR and Sm@rt-TAN photo). You can check the [list of supported (German) banks](https://shop.reiner-sct.com/tan-generatoren-fuer-sicheres-online-banking/tanjack-deluxe?locale=en) on the vendor's page.

## Summary

So is it worth it?

The device is double the price of the previously presented Molto-2 - around 80 Euro currently. On the other hand, its convenience and added security make this a much better device to keep your MFA as secure as possible. It is about twice the size of the Molto-2, though, so it is a bit of a hassle if you want to carry it around.

In my case, the Reiner SCT tanJack Deluxe has replaced my other means of MFA code generation. As suggested in my post of the Molto-2 device, I do keep backup data on an IronKey if this device ever fails or gets lost.

If you want to secure your most valuable accounts, like AWS root access, this device is probably a better alternative to individual tokens or other devices.

