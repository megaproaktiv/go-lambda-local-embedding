---
author: "Gernot Glawe"
title: "tRick: simple network 1 - Abstraktion und LoC"
date: 2019-05-20
linktitle: trick-simple-1
image: "img/2019/05/trick-overview.png"
thumbnail: "img/2019/05/trick-overview.png"
keywords: ["infrastructure as code"]
aliases:
    - /2019/trick-simple-1
toc: true
draft: false
tags: [devops,cloudformation,cicd,trick,cdk,iac]
categories: [AWS]
---

## Vergleich Infrastructure as Code (IaC) Frameworks - tRick

Ein Toolvergleich für Infrastructure as Code.

Um effektiv AWS oder generell Cloud Ressourcen zu erzeugen, verwendet man zur Erhöhung des Automatisierungsgrades "Infrastracture as Code", d.h. die Server, Datenbanken usw. werden in einer Sprache kodiert. Dieser Vorgang wird sinvollerweise über ein Framework, welches Tools dafür zur Verfügung stellt unterstützt.

Aber für welches Tool/Framework entscheidet man sich?

Hier werden wir mit Dimensionen für den tRick Benchmark Entscheidungshilfen geben.

<!--more-->

### Alle Posts


1. [Abstraktion und Lines of Code](/2019/05/trick-simple-network-1-abstraktion-und-loc.html)
2. [Geschwindigkeit](/2019/05/trick-simple-network-2-geschwindigkeit.html)
3. [Diversity (polyglott), Tooling, Fazit](/2019/05/trick-simple-network-3-diversity-polyglott-tooling-fazit.html)


---

Dabei unterscheiden sich die Frameworks in den grundlegenden Konzepten und Kriterien. Aus den unterschiedlichen Konzepten ergeben sich Features, die sich dann für die Nutzung auswirken. Die Bewertung erfolgt hier nach den Nutzungskriterien, da diese für die Auswahl eines Frameworks normalerweise herangezogen werden sollten.

Was meine ich damit? Naja - man wählt auch schon mal ein Tool aus, weil z.B. die Programmiersprache toll ist (wäre bei mir "ist in go"), oder weil man den Verfasser einfach toll findet (wäre bei mir bei z.B. [Yan Cui](https://theburningmonk.com/) ).

Tatsächlich würde man sich Studien zufolge mit so einer gefühlsmäßigen Entscheidung besser *fühlen* ([Feeler vs Thinker](http://changingminds.org/explanations/preferences/thinking_feeling.htm)), für Frameworkentscheidungen sollte man auch der Nachvollziehbarkeit halber versuchen, möglichst objektiver Kriterien zu finden. Diese können dann in einer [Entscheidungsmatrix](https://komfortzonen.de/entscheidungs-matrix-template-vorlage/) noch gewichtet werden  und dann ergibt sich der passendere Kandidat.

Was auch vorkommt ist, dass man die Gewichtungen so lange anpasst, bis der gefühlte Lieblingskandidat auch in der objektiven Entscheidung passt :).

Zurück zu möglichst objektiven Vergleichskriterien: Zuerst werden ich kurz die grundsätzlichen Unterschiede in den Konzepten beschreiben, um nach einer Definition des Lebenszyklus direkt in den Vergleich der Features zu gehen.

## Konzepte

### Aufruf der Erzeugung

Grundsätzlich unterscheiden sich die Frameworks/Tools darin, ob sie zuerst CloudFormation erzeugen, oder direkt über die AWS API gehen, um Ressource wie EC2 Instanzen zu erzeugen. Erzeugt man eine Resource direkt über die API, so muss man sich auch um das Update und Delete selber kümmern. In Cloudformation werden mit dem Stack alle Ressourcen gelöscht.

### Eigene Sprache / Programmiersprache

Die Unterscheidung der Art und Weise in der man den Resourcenzielzustand erreicht ist entweder _deklarativ_ oder _imperativ_. CloudFormation selber ist deklarativ, d.h. es wird ein Zielzustand beschrieben, der durch das Tool erreicht werden soll.
Imperativ wäre eine Aneinanderkettung von Befehlen, um die Ressourcen zu erzeugen

### Speicherung des Status

Wenn aus der Beschreibung die Ressourcen erzeugt werden, so werden z.B. für die einzelnen Instanzen Kennzeichner (IDs) erzeugt. Alle diese IDs zusammen, ihre Verbindungen untereinander und diverse Metadaten zusammen bilden den Stack. Mit diesem Stack kann man daher alle Ressourcen zusammen managen. Sozusagen alles oder nichts. Dieser Zustand des Stacks wird gespeichert, damit ich später mit einer geänderten Beschreibung den Zustand abändern kann. Wenn ich z.B. den Typ einer EC2 Instanz ändere muss ich wissen, welche ID die EC2 Instanz hat.

CloudFormation basierte Tools verlassen sich auf AWS, um den Zustand zu speichern. Für die beiden Terreform basierten Tools Terraform und Pulimi muss man selber die Speicherung des Zustands übernehmen, bzw. das Tool.

Um die Tools/Frameworks zu vergleichen, beschreibe ich zuerst inen Lebenszyklus , in den dann später z.B. die Geschwindigkeitstest gelegt werden

## Lebenszyklus

Die Aufrufe des Zyklus ist soweit möglich ein einem Makefile beschrieben, so dass man jeweils mit `make $cycle_name` für die jeweiligen Frameworks der Aufruf verglichen werden kann. Mein Ziel ist es möglichst eine vollautomatische Geschwindigkeitsmessung hinzubekommen. Schließlich geht es um Automatisierung!

1. Installation
   Da die einmalige Installation gegenüber dem mehrmaligen Lebenslauf Entwicklung bis Löschen nicht so stark ins Gewicht fällt, wird sie hier nicht weiter betrachtet.
1. Globale Initialisierung `make init`
   Hier wird das Tool an sich initialisiert. Die Ausführung ist nur einmal pro Workstation.
1. Projekt Initialisierung
   Diese Phase wird in der Betrachtung auch übersprungen, wir haben vorhandene Projekte.
1. Entwicklung
   Hier kommt das Tooling zum Tragen, um effektiv Entwickel zu können. Diese Phase wird vornehmlich im Tooling betrachtet.
1. Build - Erzeugung der deploybaren Dateien
1. Check - Änderungen prüfen
1. Deploy - Aufbau der Ressourcen
1. Löschen

Für die einzelnen Tools ergibt sich für die Phasen:

|Tool           | pre_build |build | post_build | deploy | remove|
|---            | ---   |---| ---   | ---    | --- |
|cdk            | tsc   || cdk diff | cdk deploy |cdk remove | 
|Cloudformation || Editor | Change Set | aws cloudformation create-stack | aws cloudformation delete-stack |
|pulumi         | npm install |pulumi stack select  |  |  pulumi up --yes | pulumi destroy && pulumi stack rm |
| Terraform     | |tf plan | tf plan | tf apply | tf destroy |
| Goformation|  |go build | - | wie CloudFormation| wie Cloudformation |

Nun zu dem Benchmark Dimensionen:

## Benchmarking Dimensionen

Um die Frameworks zu vergleichen, werden sie jeweils gegeneinander in der Dimension verglichen.

### Abstraktionslevel

Wieviele Resourcen werden automatisch erzeugt? Wer muss sich um die Verbindungen zwischen den einzelnen Ressourcen kümmern?

### Speed

Ausführungsgeschwindigkeit des Tools und der Erstellund der Ressourcen.
Bei der Erstellung der Ressourcen gibt es generell die Möglichkeiten erst CLoudformation zu erzeugen oder direkt API Aufrufe durchzuführen.

Die Erstellung von Cloudformation und dann Ausführung durch den AWS Service Cloudformation ist zwar tendentiell in der Erzeugung der Ressourcen leicht langsamer (wir werden es messen), hat aber den mehrere Vorteile:

- Der Zustand (State) der Ressourcen wird "as a Service" bei AWS gespeichert. Bei Terraform und Pulumi benötigt man einen State Provider
- Events aus dem Stack: Es gibt die Möglichkeit, sich in die Events bei Cloudformation direkt bei AWS einzuklinken, z.B. Jedesmal, wenn ein Stack fertig ist eine Benachrichtigung per SNS auszulösen 
- Rückwärtskompatibilität - Will man doch (was seltener vorkommt als man glaubt) das Tool wechseln, so kann man auf das generierte Cloudformation zurückgreifen
- Verwendung des CloudFormations in anderen Services, z.B. ServiceCatalog
- Verteilung des CloudFormations als simple Ein Klick Deployment Lösung

Es spricht also vieles für die Verwendung eines Frameworks, welches direkt CloudFormation erzeugt.

### Diversity

Wieviele Programmier Sprachen werden unterstützt?
Hier wird zuerst die Summe aller Sprachen aller Frameworks gezählt und dann jeweils für das einzelne Framework gezählt, wieviele der Sprachen unterstützt werden.
Das ist erstmal nur eine Eigenschaft, ob diese für die eigene Auswahl relevant ist wird dadurch nicht ausgesagt.

Wenn man z.B. immer mit JavaScript/TypeScript arbeitet, ist es evt. nicht relevant, ob Python unterstützt wird. Wenn das Framework als Standard im größeren Kontext, z.B. firmenweit verwendet werden soll, kann es sehr wohl relevant sein, wenn ein anderes Team eine andere Sprache bevorzugt.

### Tooling

Auch hier werden - hier ist es subjektiv - Features aller Frameworks zusammengezählt und jeweils geprüft, was durch das einzelne Frameworkk unterstützt wird.

### LoC

Lines of Code

Hier werden nur die Zeilen der eigentlichen Erstellung gezählt, keine Standardkonfiguration, die erzeugt wird.

Nach der Definition wenden wir den Benchmark auf simple vpc an.

---

Hier im ersten Teil betrachten wir den Abstraktionslevel und die Anzahl Zeilen (Lines of Code - LoC) als Dimensionen.

Hier ist eine gewisse Korrelation zu erwarten...

Für jede Dimension gibt es 1 ... 5 Punkte zu erreichen.

## Benchmark für die Dimension Abstraktionslevel

### Abstraktion 5 CDK

```js
    const tRickVPC = new VpcNetwork(this, 'tRick-simple-network-vpc', {
      cidr: '10.0.0.0/16',
      maxAZs: 1,
      subnetConfiguration: [
        {
          name: 'Public',
          subnetType: SubnetType.Public,
          cidrMask: 24
        }]
    });
```

Dieser Codeabschnitt erzeugt folgende Ressourcen:

```bash

Resources
[+] AWS::EC2::VPC tRick-simple-network-vpc tRicksimplenetworkvpcF740EA4A
[+] AWS::EC2::Subnet tRick-simple-network-vpc/PublicSubnet1/Subnet tRicksimplenetworkvpcPublicSubnet1Subnet51CE8708
[+] AWS::EC2::RouteTable tRick-simple-network-vpc/PublicSubnet1/RouteTable tRicksimplenetworkvpcPublicSubnet1RouteTableB41D2434
[+] AWS::EC2::SubnetRouteTableAssociation tRick-simple-network-vpc/PublicSubnet1/RouteTableAssociation tRicksimplenetworkvpcPublicSubnet1RouteTableAssociationD56BA32F
[+] AWS::EC2::Route tRick-simple-network-vpc/PublicSubnet1/DefaultRoute tRicksimplenetworkvpcPublicSubnet1DefaultRoute2118CE19
[+] AWS::EC2::InternetGateway tRick-simple-network-vpc/IGW tRicksimplenetworkvpcIGWA9D38D6B
[+] AWS::EC2::VPCGatewayAttachment tRick-simple-network-vpc/VPCGW tRicksimplenetworkvpcVPCGWFAC23D9C
```

Das CDK baut hier die sowieso notwendigen zum VPC zusätzlichen Ressourcen automatisch auf.
Z.B. die `RouteTable` wird nicht explizit definiert. Daher ist das Abstraktionslevel von den betrachteten Frameworks am höchsten.

### Abstraktion 4 Pulumi

```js
let vpc = new aws.ec2.Vpc("vpc", {
  cidrBlock: variable.cidr_block,

  tags: variable.tags
});

let public_subnet = new aws.ec2.Subnet("public_subnet", {
  availabilityZone: variable.region + variable.availability_zone,
  cidrBlock: variable.subnet_cidr_block,
  mapPublicIpOnLaunch: true,
  vpcId: vpc.id,

  tags: variable.tags
});
```

Pulumi liegt hier vor Terraform, aufgrund des Referenzproblems.
Die Referenz vom Subnet auf das VPC kann hier im Typescript definiert werden. Das ermöglicht IDE Unterstützung und ist weniger fehleranfällig als die explizite Referenziert über reinen Text.

### Abstraktion 3 Terraform

```hcl
resource "aws_vpc" "vpc" {
  cidr_block = "${var.cidr_block}"

  tags = "${var.tags}"
}

resource "aws_subnet" "public_subnet" {
  availability_zone       = "${var.region}${var.availability_zone}"
  cidr_block              = "${cidrsubnet(var.cidr_block, 8, 0)}"
  map_public_ip_on_launch = true
  vpc_id                  = "${aws_vpc.vpc.id}"

  tags = "${var.tags}"
}
```

Ich kenne zur Zeit kein Tool, was bei Terraforms eigener Sprache "hcl" bei der Referenzierung unterstützt. Durch das properitäre Format wird die Unterstützung auch schwieriger zu implementieren als z.B. bei TypeScript, welches eine breite Tooling Unterstützung geniest.

### Abstraktion 2 Goformation

```go
func buildVpc(template *cloudformation.Template, cidr string){
    var cloudformationVPC cloudformation.AWSEC2VPC
    cloudformationVPC.CidrBlock = cidr
    template.Resources[vpc] = &cloudformationVPC
}

func buildSubnet(template *cloudformation.Template, vpc string, cidr string, zone string){
    var cloudformationSubnet cloudformation.AWSEC2Subnet
    cloudformationSubnet.AvailabilityZone = zone
    cloudformationSubnet.CidrBlock = cidr
    cloudformationSubnet.VpcId = cloudformation.Ref(vpc)
    template.Resources[subnet1] = &cloudformationSubnet
}
```

Goformation unterstützt alle CloudFormation Ressourcen, da es Code aus den Beschreibungen von Cloudformation generiert. Das ist für die Abdeckung der Ressourcen von Vorteil, resultiert aber in generischem Code mit geringem Abstraktionslevel. Was aber durch die Sprache und das Tooling unterstütz wird, sind die Referenzen.

### Abstraktion 1 Cloudformation

Als Definiton des Basislevels gibt es hier nur einen Punkt.
CloudFormation selber bietet gegenüber derdirekten  Verwendung der AWS API oder CLI zur Erstellung von Resourcen bereits eine Abstraktion.

### Zusammenfassung Dimension Abstraktion

![Abstraction](/img/2019/05/trick-summary-abstraction.png)

## Lines Of Code

Ich erwarte hier natürlich eine Korrelation zum Abstraktionslevel...

Ein paar Kommentare fallen nicht so ins Gewicht, der Trend ist wichtig.

#### LoC Terraform

|Datei | Inhalt | Loc |
| --- | --- | ---|
| main.tf |Alles | 65  |
| variables.tf |Alles | 25   |
| Summe ||  90  |

#### LoC Pulumi

|Datei | Inhalt | Loc |
| --- | --- | ---|
| index.ts |Alles | 62 |
| Summe || 62 |

#### LoC GoFormation

|Datei | Inhalt | Loc |
| --- | --- | ---|
| main.go |Alles | 97 |
| Summe || 97 |

#### LoC CDK

|Datei | Inhalt | Loc |
| --- | --- | ---|
| bin/cdk.ts | Stack selber | 10 |
| lib/cdk-stack.t | VPC Definition |30 |
| Summe || 40 |

#### LoC CloudFormation

|Datei | Inhalt | Loc |
| --- | --- | ---|
| template.json |Alles | 74 |
| Summe || 74 |

### Zusammenfassung Dimension LoC

![Loc](/img/2019/05/trick-summary-loc.png)

Hier sind nur die Anzahl der Zeilen zu sehen, die Punkte sind natürlich genau anders herum, also wenig Zeilen, viele Punkte.

## Fazit Teil 1

Für die Erstellung und Verwaltung von vielen Ressourcen ist der Abstraktionsleven enorm wichtig. Daher ist für mich persönlich das CDK eine sehr gute Wahl für ein IaC Framework.