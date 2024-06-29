---
title: "How to use hugo"
author: "Gernot Glawe"
date: 2024-02-11
draft: false
image: "shraga-kopstein-Dn2eTdtP7jA-unsplash.jpg"
thumbnail: "shraga-kopstein-Dn2eTdtP7jA-unsplash.jpg"
toc: true
keywords:
    - documents
    - pow
    - markdown
tags:
    - level-100
    - hugo
    - markdown
categories: [editors]
---

# How to create an article in hugo

Work these steps in the pow repository:

1) Create a new directory for the year and the post

E.g.: `content/post/2024/my-first-post`

2) Add frontmatter to the markdown file


**Frontmatter**

```markdown
---
title: "Authenticate local react app with static AWS credentials"
author: "Gernot Glawe"
date: 2024-01-11
draft: false
image: "poc2prod2.jpg"
thumbnail: "poc2prod2.jpg"
toc: true
tags:
    - level-300
    - LLM
    - genai
    - llm
categories: [aws]
---
```

Example to copy:

```markdown
---
title: "<A short headline for my problem solutions>"
author: "<John Doe>"
date: <2024-01-11>
draft: true
image: "poc2prod2.jpg"
thumbnail: "poc2prod2.jpg"
toc: true
tags:
    - level-<300>
    - <awsservice>
    - <othertechnologi>
    - llm
categories: [aws]
---

# Problem

# Solution

# Prequisites

# Summary

# See also

# Thanks to


3) Add images to the directory `content/post/2024/my-first-post`

Sources:
- https://unsplash.com/
- https://thenounproject.com/

** ONLY USED LICENSE FREE IMAGES **
** Add shoutout to author in the post (Part Thanks to)**

- Scale down images to aprox 1600px width
    - Preview mac can do this
      - Open image in preview
        - Tools -> Adjust size (Werkzeuge/Größenkorrektur)
            - Adjust width % (Breite prozent)  to 1600px or below
            - Size should be below 1MB
            ![Preview](adjust-image.png)

4) Add content to the markdown file
5) add tags, so [updatealtert](https://github.com/megaproaktiv/updatealert) can find them

6) add categories like:
    - development
    - projectorganisation

Categories are used for very broad knowledge areas, tags are used for more specific topics. Its more like "what roles are interested in this" vs. "what is the topic".

7) Start with setting 'draft: true'
8) Run `git submodule update --init --recursive`
9) Check the result with `hugo server` and open the browser at `http://localhost:1313/`
10) When satisfied, set `draft: false` and push to the repository

For automatic translation, embedd all words which should not be translated with `backticks`:

```md
This post is about `AWS` and `React`
```
