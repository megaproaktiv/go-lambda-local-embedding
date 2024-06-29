---
author: "Gernot Glawe"
title: "tRick: simple network 2 - Geschwindigkeit"
date: 2019-05-30
linktitle: trick-simple-2
image: "img/2019/05/trick-overview.png"
thumbnail: "img/2019/05/trick-overview.png"
aliases:
    - /2019/trick-simple-2
toc: true
draft: false
tags: [devops,cloudformation,cicd,trick,iac]
categories: [AWS]
---

## Vergleich Infrastructure as Code (IaC) Frameworks - tRick

### Alle Posts


1. [Abstraktion und Lines of Code](/2019/05/trick-simple-network-1-abstraktion-und-loc.html)
2. [Geschwindigkeit](/2019/05/trick-simple-network-2-geschwindigkeit.html)
3. [Diversity (polyglott), Tooling, Fazit](/2019/05/trick-simple-network-3-diversity-polyglott-tooling-fazit.html)



---

## Benchmark Ausführungsgeschindigkeit


### Ausführungsgeschwindigkeit

Direkt aus dem tRick Repository wird mehrfach (n=10) der Zyklus Build -> Check -> Deploy -> Remove ausgeführt. Damit sollen Cache Effekte statistisch gemittelt werden. Dazu nehme ich das Tool `hyperfine` zur Hilfe. Es führt Kommandos automatisch mehrfach aus und mittelt die Ergebnisse.

Meine Annahme ist es, dass Terraform vorne liegt, da das Programm selber statisch kompiliert in go geschrieben ist. Außerdem geht die Ausführung direkt auf die API.

<!--more-->

Als Sprache bei Pulumi und CDK wird Typscript verwendet.

Generell würden mich persönlich Differenzen in der Ausführungszeit bis ca 3 Sekunden nicht stören, aber erst mal messen!

#### Geschwindigkeit CDK

| Command | Mean [s] | Min…Max [s] |
|:---|---:|---:|
| `../speed/cycle.sh` | 135.861 ± 3.661 | 130.339…141.943 |

Die Funktionalität fordert hier ihren (Geschwindigkeits) Preis. Zu beachten ist, dass hier immer vollständige Durchgänge gemessen werden. Im der Entwicklung sollten eher Updates der Fall sein. Wenn der Anwendungsfall aber z.B. ein schnelles Aufbauen von Kundenportalen ist, währende der Kunde darauf wartet, könnte sich die Verwendung von reinem CloudFormation lohnen. Das bekommt man aus dem CDK mit `cdk synth` ausgegeben.

#### Geschwindigkeit Terraform

| Command | Mean [s] | Min…Max [s] |
|:---|---:|---:|
| `../speed/cycle.sh` | 29.116 ± 1.992 | 28.131…34.605 |

#### Geschwindigkeit Pulumi

Für Pulumi muss man sich von Anfang an beim pulumi Server anmelden.
Eine vollautomatischer Auf und Abbau habe ich nach einiger Ausprobiererei gelassen, daher nehme ich hier die einmaligen manuellen Werte:

```bash
time pulumi stack init ****/simplevpc
#real	0m1.641s
time pulumi  up -y
# real	0m19.977s
time pulumi destroy -y
# real	0m22.958s
```

Also ca. 44,57 Sekunden.

#### Geschwindigkeit GoFormation

Auslieferung des Stacks mit dem Tool Clouds.

| Command | Mean [s] | Min…Max [s] |
|:---|---:|---:|
| ` ../speed/cycle.sh goformation` | 95.809 ± 4.242 | 90.154…106.028 |

Hier ist zu beachten, dass die Kompilierung des go Programms dazu gezählt wird. Die reine Generierung ist sehr schnell.

#### Geschwindigkeit CloudFormation

| Command | Mean [s] | Min…Max [s] |
|:---|---:|---:|
| `../speed/cycle.sh` | 17.180 ± 25.861 | 4.460…66.312 |

Hier nehme ich die reine AWS CLI zum Deployen.

### Zusammenfassung 

In der Zusammenfassung hier die Ergebnisse:

![Speed](/img/2019/05/trick-speed.png)

Auch hier gibt natürlich die kürzeste Ausführungszeit am meisten Sterne.


Erstaunlich finde ich, dass tatsächlich reines CloudFormation - jedenfalls in diesem UseCase - noch vor Terraform liegt! DIe Optimierung der Parallelität scheint direkt bei AWS noch besser zu gehen.
Außerdem hat CloudFormation den Vorteil, dass die endgültigen API Calls direkt von AWS zu AWS gehen und nicht von der Workstation zu AWS. Verwendet man terraform in einer **CodeBuild** pipeline, so entfällt dieser Vorteil.
 
Damit ergibt sich für 3 von 5 Dimensionen folgendes Übersichtsbild:

{{< figure src="/img/2019/05/trick-overview-2.png" title="tRick 3 Dimensionen von 5 ermittelt" >}}

Damit macht das "Spinnen" Bild am Anfang des Artikels auch langsam Sinn... 

Die restlichen zwei Dimensionen, nämlich "Diversity" und "Tooling" folgen im dritten Teil.