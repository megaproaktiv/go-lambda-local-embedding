---
author: "Maurice Borgmeier, Alexey Vidanov"
title: "Aufbau von Lambda mit terraform"
date: 2024-03-14
image: "img/2019/05/terraform.png"
thumbnail: "img/2019/05/terraform.png"
toc: true
draft: false
tags: [devops,terraform,lambda,iac]
categories: [AWS]
---

{{% notice note %}}
Note: This is an updated version of [this blog](https://www.tecracer.com/blog/2019/05/building-lambda-with-terraform.html).
{{% /notice %}}

# Aufbau von Lambda Funktionen mit Terraform

## Einleitung

Vielfach wird `terraform` verwendet, um die AWS Ressourcen als Code "Infrastructure as Code" zu managen.

Für uns als AWS Benutzer wird Lambda immer mehr zu einem wichtigen Teil der Infrastruktur und vor allem deren Automatisierung. Das Ausrollen und speziell das Erstellen/Kompilieren (build) von Lambda Funktionen mit Terraform geht leider nicht ganz so einfach.

Um fair zu bleiben - es lässt sich darüber streiten, ob man Terraform für diesen Anwendungsfall überhaupt nutzen sollte. Ich möchte genau das tun, sonst gäbe es auch diesen Blog Eintrag nicht, also weiter im Text.

Den Ablauf zum Deployen einer Lambda Funktion habe ich hier in drei Schritte unterteilt: **Build**, **Compress** und **Use** - die werde ich gleich kurz vorstellen.

Des Weiteren bezeichne ich das ab jetzt als  Build Pipeline, auch wenn Terraform nicht gerade ein Build Tool ist - versucht mich aufzuhalten :)

Zum Schluss zeige ich zwei Beispiele für Lambda Build-Pipelines in Terraform:

1. Eine vereinfachte Version nur mit den Schritten "compress" und "use"
2. .. und die komplexe Version mit allen drei Build-Schritten


### Build

Dieser Schritt kann viele verschiedene Dinge beinhalten, das hängt wie immer von dem Anwendungsfall und der Laufzeitumgebung ab:

- Installation von Abhängigkeiten (z.B. Python Paketen)
- Testausführung
- Code Kompilation
- Konfiguration
- ... und so weiter

Wir nehmen einfach mal an, dass wir ein Script ausführen und bei dem "kein Fehler" Rückgabewert `0` weitermachen.

### Compress

Um Lambda Funktionen zu deployen, kann man zwar auch Inline Lambda Code verwenden, aber der Normalfall ist es, eine zip Datei (daher compress) zu erzeugen.
Deswegen müssen wir in unserer Pipeline ein Zip Archiv erstellen.

### Use

Dieses Zip Archiv verwenden wir dann um die Lambda Funktion zu erzeugen.

## Beispiel 1 - Vereinfachte Build Pipeline

In dieser vereinfachten Pipeline überspringen wir den "build" Schritt, den brauchen wir nur, wenn wir Pakete verwenden möchten, die nicht Teil der Standard-Laufzeitumgebung sind.

Die Verzeichnisstruktur des Beispielprojektes sieht wie folgt aus:

<!--
    Command for the tree-view
    tree -I 'venv|environments|switch_environment.sh|*.md|*.zip'
-->

```text
├── code
│   └── my_lambda_function
│       └── handler.py
├── lambda.tf
├── main.tf
├── permissions.tf
└── variables.tf

```

<!-- We're going to skip the IAM Role, because that's not very interesting --> 

Diese Terraform-Ressource stellt den *Compress* Schritt da - wir nutzen hier die `archive_file` [Data Source](https://www.terraform.io/docs/providers/archive/d/archive_file.html)
des Terraform Archive-Providers (Bei der ersten Verwendung in einem Projekt muss anschließend mit `terraform init` der neue Provider initialisiert werden).

Wo das komprimierte Zip-Archiv gespeichert wird (`output_path`) ist nicht wirklich wichtig - es lohnt sich aber in jedem Fall, das Archiv in die
`.gitignore` Datei aufzunehmen, denn sowohl den Code als auch das Build-Artefakt einzuchecken ist nicht notwendig. 

```hcl-terraform
data "archive_file" "my_lambda_function" {
  source_dir  = "${path.module}/code/my_lambda_function/"
  output_path = "${path.module}/code/my_lambda_function.zip"
  type        = "zip"
}
```

Jetzt können wir mit dem nächsten Schritt weitermachen. In der Funktionsdefinition wird eine IAM-Rolle referenziert, die hier nicht dargestellt 
ist - hier solltet ihr eure eigene verwenden. Der `filename` parameter zeigt auf die oben erwähnte Data Source - unser komprimiertes Build-Artefakt.
Der `source_code_hash` Parameter referenziert den SHA-256 Hash des Build-Artefakts und sorgt im Kern dafür, dass der Code der Lambda-Funktion nur
ausgetauscht wird, wenn sich der Hash ändert - sprich: wenn sich der Code ändert.

```hcl-terraform
resource "aws_lambda_function" "my_lambda_function" {
  function_name    = "my_lambda_function"
  handler          = "handler.lambda_handler"
  role             = "${aws_iam_role.my_lambda_function_role.arn}"
  runtime          = "python3.11"
  timeout          = 60
  filename         = "${data.archive_file.my_lambda_function.output_path}"
  source_code_hash = "${data.archive_file.my_lambda_function.output_base64sha256}"
}
```

Das war's auch schon - nachdem `terraform apply` ausgeführt wurde solltet ihr in der Konsole den aktuellen Code sehen (bei Änderungen am Code
dauert es manchmal ein paar Sekunden, bis diese in der Konsole sichtbar sind).

## Beispiel 2 - Vollständige Version der Build Pipeline

Unsere Ordnerstruktur sieht jetzt wie folgt aus - vielleicht könnt ihr Gemeinsamkeiten erkennen...:
```
├── code
│   └── my_lambda_function_with_dependencies
│       ├── build.sh
│       ├── handler.py
│       ├── package
│       └── requirements.txt
├── lambda.tf
├── main.tf
├── permissions.tf
└── variables.tf
```

Das `build.sh` Shell-Script ist relativ simpel, aber effektiv. Es navigiert zunächst zum Speicherort des Scriptes und installiert dann die Abhängigkeiten aus der `requirements.txt` in den `package` Ordner.

```
#!/usr/bin/env bash

# Change to the script directory
cd "$(dirname "$0")"
pip install -r requirements.txt -t package/
```

Die `handler.py` sieht wie folgt aus - das Script nutzt das `requests` Modul um die öffentliche IP der Lambda-Funktion (oder eines Proxies) zu ermitteln:

```python
# Tell python to include the package directory
import sys
sys.path.insert(0, 'package/')

import requests

def lambda_handler(event, context):

    my_ip = requests.get("https://api.ipify.org?format=json").json()

    return {"Public Ip": my_ip["ip"]}

```

Weiter geht es mit der Build-Pipeline.

Der eigentliche Build-Schritt ist eine Null-Ressource. Sie führt über den `local-exec` Provisioner das Build-Script aus, wenn sich an einer der folgenden Dateien etwas geändert hat:
- `handler.py`
- `requirements.txt`
- `build.sh`

```hcl-terraform
resource "terraform_data" "my_lambda_buildstep" {
  triggers_replace = {
    handler      = "${base64sha256(file("code/my_lambda_function_with_dependencies/handler.py"))}"
    requirements = "${base64sha256(file("code/my_lambda_function_with_dependencies/requirements.txt"))}"
    build        = "${base64sha256(file("code/my_lambda_function_with_dependencies/build.sh"))}"
  }

  provisioner "local-exec" {
    command = "${path.module}/code/my_lambda_function_with_dependencies/build.sh"
  }
}
```

Der Compress-Schritt sieht fast genau so aus, wie oben, mit Ausnahme der `depends_on` Anweisung. Hier sagen wir Terraform, dass es bitte warten soll, bis der Build-Schritt abgeschlossen ist, bevor das Ergebnis komprimiert wird.

```hcl-terraform
data "archive_file" "my_lambda_function_with_dependencies" {
  source_dir  = "${path.module}/code/my_lambda_function_with_dependencies/"
  output_path = "${path.module}/code/my_lambda_function_with_dependencies.zip"
  type        = "zip"

  depends_on = ["terraform_data.my_lambda_buildstep"]
}
```

Abschließend verwenden wir das entstehende Build-Artefakt - wie oben - um die Lambda-Funktion zu definieren.

```hcl-terraform
resource "aws_lambda_function" "my_lambda_function_with_dependencies" {
  function_name    = "my_lambda_function_with_dependencies"
  handler          = "handler.lambda_handler"
  role             = "${aws_iam_role.my_lambda_function_role.arn}"
  runtime          = "python3.11"
  timeout          = 60
  filename         = "${data.archive_file.my_lambda_function_with_dependencies.output_path}"
  source_code_hash = "${data.archive_file.my_lambda_function_with_dependencies.output_base64sha256}"
}
```

Das war's - nach einem `terraform apply` sollte das Ergebnis nach wenigen Sekunden in der Konsole zu sehen sein.

## Sonstiges

Diese Lösung ist von einer [Diskussion auf Github](https://github.com/hashicorp/terraform/issues/8344) inspiriert, 
danke an [@dkniffin](https://github.com/dkniffin) und [@pecigonzalo](https://github.com/pecigonzalo)  

