---
author: "Thomas Heinen"
title: "VSCode Repository-Level Task Definitions"
date: 2022-11-25
image: "img/2022/11/annie-vo-HUuWNQmW58A-unsplash.png"
thumbnail: "img/2022/11/annie-vo-HUuWNQmW58A-unsplash.png"
toc: false
draft: false
tags:
  - allgemein
  - cli
  - devops
  - developing
  - vscode
  - iac
---
Do you run the same CLI commands again and again while using VSCode? Even if you already put them into code, you find yourself typing things like `rake build` all the time?

I just learned of VSCode's integrated Task management the other day, and this knowledge could help you work more productively. So let's dive deep...

<!--more-->

During my regular workday, I switch between different types of repositories. Some are Rubygems, some are related to Terraform/Terragrunt, and some are even for Go.

I was largely content with remembering commands or writing a `Rakefile` until I worked on streamlining a tool for colleagues who might only use GUI-level features usually. That got me thinking if there might be a better way.

## VSCode Workspace Tasks

It turns out VSCode has [Workspace Tasks](https://code.visualstudio.com/docs/editor/tasks) integrated. 

These tasks allow you to configure [custom, repository-specific tasks](https://code.visualstudio.com/docs/editor/tasks#_custom-tasks)[^1] and even get input from the user. The feature works via a `tasks.json` file like this one:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "My repetitive task",
      "type": "shell",
      "command": "execute-some-tool --with=argument",
      "options": {
        "env": {
          "DEBUG": 1
        }
      },
      "problemMatcher": []
    }
  ]
}
```

You put this into your repository in the `.vscode/` directory. VSCode will instantly pick up changes, and you can access and execute these tasks.

For this, use the quick open function (`Ctrl+P`) and type `task ` to give a dropdown of all configured/detected tasks.

![Task Dropdown](/img/2022/11/vscode-tasks-tasks.png)

You can even create inputs for your tasks, so you might offer specific preselected values or accept free-text answers:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "My repetitive task",
      "type": "shell",
      "command": "execute-some-tool --with=${input:format}",
      "problemMatcher": []
    }
  ],
  "inputs": [
    {
      "type": "pickString",
      "id": "format",
      "description": "What format to export?",
      "options": ["docx", "md", "json"],
      "default": "docx"
    }
  ]
}   
```

![Task Inputs](/img/2022/11/vscode-tasks-inputs.png)

If you think getting to these tasks is too cumbersome, we have the same opinion. Luckily there are two built-in shortcuts to this.

You can configure one of the tasks to be the default "build" task (that is if your project does not have a Build step). Just assign it to the `build` group and set it as default to get access via `Ctrl+Shift+B`:

```json
      "group": {                                                                 
        "kind": "build",                                                         
        "isDefault": true                                                        
      }
```

Or, you can go the more straightforward way and assign hotkeys via `keybindings.json` (inside `.vscode/`):

```json
[
  {
    "key": "ctrl+h",
    "command": "workbench.action.tasks.runTask",
    "args": "My repetitive task"
  }
]
```

## Extension "Tasks": Statusbar Shortcuts

When we venture into extension land, we can also find better ways to access our tasks.

The [Tasks extension](https://marketplace.visualstudio.com/items?itemName=actboy168.tasks) gives you more `options` inside your JSON. In particular, adding the most important tasks right into the status bar, coloring them, or using icons:

![Statusbar Integration](/img/2022/11/vscode-tasks-statusbar.png)

## Extension "Task Runner": Task overview

The [Task Runner extension](https://marketplace.visualstudio.com/items?itemName=SanaAjani.taskrunnercode) adds a "Task Runner" pane to your Explorer side pane. This pane will display all configured tasks, whether auto-detected ones or workspace tasks. 

![Task Runner Pane](/img/2022/11/vscode-tasks-taskrunner.png#center)

If you cannot see the Task Runner pane, click the three dots left of "Explorer" and switch it on.

## Auto-Detected Tasks

In some cases, you might be lucky that you do not need to configure your own `tasks.json`, because you use a language or task runner that is natively supported (Grunt, Gulp, Jake, NPM, ...)

Using codified tasks aside from `tasks.json` gives us an important feature: Reusability.

While only VSCode can use its Workspace Tasks, a different solution like Grunt, Rake, or even Make makes it possible to use these from the CLI or inside your CI/CD system. This technique combats the "Works on my machine" syndrome, as IDE and CI/CD use the same definitions.

Extensions can register custom Task Providers, so you can pick your preferred solution even if it is not supported natively by VSCode.

### Detecting Rakefile/Taskfile

For most Ruby-fan, `rake` is the tool of choice. The popular [Rebornix Ruby extension](https://marketplace.visualstudio.com/items?itemName=rebornix.Ruby) implemented auto-detecting tasks in 2017 and works out of the box. As it has no UI element, feel free to combine it with the usability-related extensions mentioned earlier.

![Rake Task (auto-detected)](/img/2022/11/vscode-tasks-rake.png)

In Golang, `task` is preferred. While this is not included in the default Go extension yet, you can add the corresponding [taskfile.dev extension](https://marketplace.visualstudio.com/items?itemName=paulvarache.vscode-taskfile), which comes with its own "Taskfile" pane.

![Taskfile Pane](/img/2022/11/vscode-tasks-taskfile.png#center)

Remember to codify your extension choices and not add them manually to your IDE. You can use the `.vscode/settings.json` if you develop locally or the `.devcontainer/devcontainer.json` file if you already work with DevContainers[^2].

## Summary

Adding extension configuration and task definitions to our repositories has several advantages: Quicker development cycles, reduced error rates, and consistent IDE/CI execution.

As such, it is a small piece of the "shift left" movement that tries to catch problems earlier in development - preferably already within the IDE.

## Footnotes

[^1]: If you like specifications, you can have a look at the [tasks.json specification](https://code.visualstudio.com/docs/editor/tasks-appendix#_schema-for-tasksjson)
[^2]: I touched on the concepts in my [blog about testing Terraform](https://www.tecracer.com/blog/2021/10/testing-terraform-with-inspec-part-2.html)
