---
title: "🇩🇪 Verbesserung der deutschen Suche im Amazon OpenSearch Service" 
author: "Alexey Vidanov" 
date: 2023-12-08
toc: true 
draft: false 
image: "img/2023/12/opensearch-improved.png" 
thumbnail: "img/2023/12/opensearch-improved.png" 
categories: ["aws"] 
tags: [ "aws", "opensearch", "level-400", "deutsche suche", "unternehmenssuche", "deutsch" ]
---

Der Amazon OpenSearch Service, der auf dem robusten OpenSearch-Framework basiert, zeichnet sich durch seine bemerkenswerte Geschwindigkeit und Effizienz in Such- und Analysefunktionen aus. Trotz seiner Stärken sind die Standardkonfigurationen des Dienstes möglicherweise nicht vollständig darauf ausgelegt, die spezifischen sprachlichen Herausforderungen bestimmter Sprachen zu bewältigen.

<!--more-->

{{% notice note %}}
For the English version of this article, please use [this link](https://www.tecracer.com/blog/2023/12/enhancing-german-search-in-amazon-opensearch-service.html).
{{% /notice %}}

Nehmen wir zum Beispiel das Deutsche, bekannt für seine zusammengesetzten Wörter wie „**Lebensversicherungsgesellschaft**“. Standardmäßige Tokenisierung in Suchtechnologien behandelt diese Zusammensetzungen als einzelne Einheiten, was zu weniger optimalen Suchergebnissen führt. Für eine verbesserte Genauigkeit ist es wichtig, die Bestandteile dieser Zusammensetzungen separat zu indizieren – „**Leben**“, „**Versicherung**“ und „**Gesellschaft**“. Dieser Ansatz stellt präzisere und effektivere Suchergebnisse sicher, insbesondere in Sprachen wie Deutsch mit vielen zusammengesetzten Wörtern.

![Verbesserung der deutschen Suche im Amazon OpenSearch Service](/img/2023/12/better-search-german.png)

<!--more-->

## Kombination traditioneller Suche mit erweiterten Filtern

Stand Dezember 2023 unterstützt OpenSearch eine Reihe von Sprachoptionen für die `analyzer`-Funktion. Diese Sprachen umfassen: `arabisch`, `armenisch`, `baskisch`, `bengalisch`, `brasilianisch`, `bulgarisch`, `katalanisch`, `tschechisch`, `dänisch`, `niederländisch`, `englisch`, `estnisch`, `finnisch`, `französisch`, `galicisch`, `deutsch`, `griechisch`, `hindi`, `ungarisch`, `indonesisch`, `irisch`, `italienisch`, `lettisch`, `litauisch`, `norwegisch`, `persisch`, `portugiesisch`, `rumänisch`, `russisch`, `sorani`, `spanisch`, `schwedisch`, `türkisch` und `thailändisch`.

Wenn man jedoch den deutschen Analyzer auf unser obiges Beispiel anwendet, wird deutlich, dass er bei zusammengesetzten Wörtern Schwierigkeiten hat, diese effektiv in einfachere, suchbare Elemente zu zerlegen.

```json
GET _analyze
{
  "analyzer":"german",
  "text": ["Lebensversicherungsgesellschaft."]
}
```

Das Ergebnis ist jedoch nur ein Token: `lebensversicherungsgesellschaft`. Der eingebaute deutsche Analyzer wandelt die Eingabe in Kleinbuchstaben um und entfernt Stoppwörter wie "und", "oder", "das", die für die Suche nicht wesentlich sind. Anschließend wird eine Stammformreduktion durchgeführt, um Wörter suchbarer zu machen. Leider wird die Komplexität der deutschen Sprache dabei nicht ausreichend berücksichtigt.

Um diese Herausforderung zu bewältigen, wenden Entwickler oft `n-grams` an. Diese Methode zerlegt den Text in kleinere Teile einer bestimmten Größe. Zum Beispiel ergibt der Satz „**Die Suche ist herausfordernd**“ bei Anwendung von 3-5 Grammen Tokens:

*die, suc, such, suche, uch, uche, che, her, hera, herau, era, erau, eraus, rau, raus, rausf, aus, ausf, ausfo, usf, usfo, usfor, sfo, sford, sforder, for, ford, forde, order, ordern, der, dern, ern*

Obwohl dies helfen kann, führt es oft zu vielen falschen Positiven. Es erzeugt zahlreiche bedeutungslose (z.B. her, ern, che) oder irreführende Tokens (z.B. ford, order, ordern), was zu aufgeblähten Indizes und erhöhter Clusterlast führt. Dies wirkt sich auf die Suchpräzision und Betriebskosten aus.

Die Integration von `dictionary_decompounder` und `synonym`-Filtern bietet einen verfeinerten Ansatz. Diese Filter erhöhen die Präzision und Effizienz bei der Verarbeitung deutscher Zusammensetzungen, indem sie diese in einfachere Tokens zerlegen. Zusätzlich erweitert die Synonymfunktionalität die Reichweite der Suche, indem sie unterschiedliche Ausdrucksformen ähnlicher Konzepte erkennt, was die Suchgenauigkeit und -umfassendheit weiter verbessert.

## Implementierung verbesserter Filter im Amazon OpenSearch

Der Prozess der Einrichtung dieser Filter ist unkompliziert und führt zu einer deutlich verbesserten Sucherfahrung. Die Dekompoundierungsfilter sind hervorragend geeignet, um komplexe Zusammensetzungen zu zerlegen, während die Synonymfilter die Suchfähigkeiten erweitern, um verschiedene Ausdrucksformen ähnlicher Konzepte einzubeziehen.

## Voraussetzungen

- **Amazon OpenSearch Service Cluster**: Sie sollten einen Amazon OpenSearch Service Cluster eingerichtet und betriebsbereit haben. Wenn Sie nicht sicher sind, wie Sie dies tun, bietet Amazon eine [umfassende Anleitung](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/create-cluster.html) dafür.
- **Zugang zur AWS Management Console**: Sie benötigen Zugang zur AWS Management Console mit den erforderlichen Berechtigungen, um den Amazon OpenSearch Service zu verwalten.

Stellen Sie sicher, dass Sie diese Voraussetzungen erfüllt haben, bevor Sie mit den Schritten zur Implementierung von Dekompoundierungs- und Synonymfiltern für die deutsche Sprachsuche fortfahren.

### Schritt 1. Beschaffung der Wörterbücher

Um Dekompoundierungsfilter effektiv zu implementieren, beginnen Sie mit der Beschaffung oder Erstellung einer Wortliste. Für eine allgemeine Textdekompoundierung im Deutschen sollten Sie die Lösung von Uwe Schindler und Björn Jacke in Betracht ziehen. Diese Lösung ist auf GitHub verfügbar und bietet eine Wortliste [hier](https://raw.githubusercontent.com/uschindler/german-decompounder/master/dictionary-de.txt) und Trennungsregeln [hier](https://github.com/uschindler/german-decompounder/blob/master/de_DR.xml). Hinweis: Um sie mit Amazon OpenSearch Service Cluster zu verwenden, entfernen Sie die zweite Zeile in der Datei, die mit `<!DOCTYPE` beginnt.

Für Synonyme nutzen Sie die Datei von Openthesaurus.de, verfügbar [hier](https://github.com/PSeitz/germansynonyms/blob/master/german.syn). Um sie für OpenSearch anzupassen, ersetzen Sie Leerzeichen durch Kommas.

Um Ihnen die Verwendung dieser Dateien zu erleichtern, habe ich sie entsprechend angepasst und im Repository hochgeladen.

1. [de_DR.xml](https://github.com/vidanov/tecracer-blog-projects/blob/main/opensearch_german_search/de_DR.xml) ist eine Datei mit Trennungsregeln.
2. [german-decompound.txt](https://github.com/vidanov/tecracer-blog-projects/blob/main/opensearch_german_search/german-decompound.txt) ist das deutsche Wörterbuch für Dekompoundierung.
3. [german_synonym.txt](https://github.com/vidanov/tecracer-blog-projects/blob/main/opensearch_german_search/german_synonym.txt) ist ein Synonymwörterbuch.

Beachten Sie Lizenzvereinbarungen, falls Sie diese Dateien verwenden möchten.

### Schritt 2: Hinzufügen von Wörterbüchern zum Amazon OpenSearch Service

![Adding dictionaries to Amazon OpenSearch Service](/img/2023/12/image-20231125190208837.png)

1. Erstellen Sie einen S3-Bucket und laden Sie die Dateien hoch.
2. Greifen Sie auf die AWS-Konsole Ihres verwalteten OpenSearch-Clusters zu.
3. Registrieren Sie Ihre Pakete über den Link "Pakete".
4. Verknüpfen Sie die Pakete mit Ihrem OpenSearch-Cluster.

Hinweis: Sie können diesen Schritt mit IaC-Tools wie Terraform oder CDK automatisieren.

### Schritt 3: Erstellung eines Index mit einem benutzerdefinierten deutschen Analyzer in OpenSearch

Nachdem Sie die notwendigen Pakete erhalten haben, können Sie diese in Ihren OpenSearch-Index integrieren, indem Sie ihre jeweiligen Paket-IDs verwenden. Stellen Sie sicher, dass Sie die Platzhalter-IDs (`FYYYYYYY` für *german_synonym.txt* und `FXXXXXXX` für *german-decompound.txt* und `FZZZZZZZ` für *de_DR.xml*) in Ihrer Implementierung durch tatsächliche ersetzen.

**Um mit der Indexerstellung fortzufahren:**

1. Öffnen Sie die OpenSearch Dashboards und navigieren Sie zum Abschnitt DevTools.
2. Geben Sie im DevTools-Konsolenfenster die folgenden Abfragen ein und führen Sie sie aus, um Ihren Index mit dem benutzerdefinierten deutschen Analyzer zu erstellen:

```json
PUT /german_index
{
  "settings": {
    "index": {
      "analysis": {
        "analyzer": {
          "german_improved": {
            "tokenizer": "standard",
            "filter": [
              "lowercase",
              "german_decompounder",
              "german_stop",
              "german_stemmer"
            ]
          },
          "german_synonyms": {
            "tokenizer": "standard",
            "filter": [
              "lowercase",
              "german_decompounder",
              "synonym",
              "german_stop",
              "german_stemmer"
            ]
          }
        },
        "filter": {
          "synonym": {
            "type": "synonym",
            "synonyms_path": "analyzers/FYYYYYYY"
          },
          "german_decompounder": {
            "type": "hyphenation_decompounder",
            "word_list_path": "analyzers/FXXXXXXX",
            "hyphenation_patterns_path": "analyzers/FZZZZZZZ",
            "only_longest_match": false,
            "min_subword_size": 3
          },
          "german_stemmer": {
            "type": "stemmer",
            "language": "light_german"
          },
          "german_stop": {
            "type": "stop",
            "stopwords": "_german_",
            "remove_trailing": false
          }
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "paragraph": {
        "type": "text",
        "analyzer": "german_improved",
        "search_analyzer": "german_synonyms"
      }
    }
  }
}
```

In dieser Einrichtung verwenden wir zwei verschiedene Analyzer: 'german_improved' für die Indizierung und 'german_synonyms' für die Suche. Dies soll sowohl die Speicher- als auch die Suche effizienz verbessern.

- **Indizierung mit dem 'german_improved' Analyzer**: Während der Indizierung wird der 'german_improved' Analyzer verwendet. Dieser Analyzer umfasst einen Standard-Tokenizer und eine Reihe von Filtern einschließlich Lowercase, german_decompounder, german_stop und german_stemmer. Das Hauptziel hier ist es, den Text zu dekomponieren und zu standardisieren, um die Konsistenz und Relevanz des Index zu verbessern. Wichtig ist, dass dieser Analyzer bewusst keinen Synonymfilter verwendet, um einen schlankeren und kompakteren Index zu erhalten, der sich auf die grundlegenden sprachlichen Elemente der deutschen Sprache konzentriert, ohne die zusätzliche Komplexität und den Speicherbedarf von Synonymen.
- **Suche mit dem 'german_synonyms' Analyzer**: Für die Suche wechseln wir zum 'german_synonyms' Analyzer. Dieser Analyzer hat dieselben Grundkomponenten wie 'german_improved', fügt jedoch eine entscheidende Ebene hinzu - den Synonymfilter. Dies erhöht die Flexibilität und Relevanz der Suche erheblich, indem eine Reihe von synonymen Begriffen berücksichtigt wird, wodurch der Suchbereich erweitert wird, ohne die Genauigkeit zu beeinträchtigen.
- **Effizienz und Leistung**: Eines der bemerkenswerten Ergebnisse dieser Methodik ist die erhebliche Reduzierung der Indexgröße – mindestens 50% kleiner im Vergleich zur Verwendung von ngrams. Das Ausmaß dieser Reduzierung kann je nach ngram-Bereich variieren. Dieser schlankere Index spart nicht nur Speicherplatz, sondern trägt auch zu schnelleren Suchvorgängen bei. Darüber hinaus gewährleistet die ausschließliche Verwendung von Synonymen für die Suchphase, dass der Index fokussiert und effizient bleibt, während die Suchvorgänge inklusiver und kontextuell bewusster werden.
- **Benutzerdefinierte Filter und Dekompoundierung**: Die benutzerdefinierten Filter wie 'german_decompounder', 'german_stemmer' und 'german_stop' sind speziell auf die einzigartigen Merkmale der deutschen Sprache zugeschnitten, wie z.B. zusammengesetzte Wörter und unterschiedliche Flexionen. Der Dekompoundierer ist insbesondere ein leistungsfähiges Werkzeug, um komplexe deutsche Zusammensetzungen in besser suchbare Elemente zu zerlegen und so sowohl die Indizierung als auch die Suchprozesse weiter zu verfeinern.

Durch den Einsatz dieser Dual-Analyzer-Strategie erreichen wir ein optimales Gleichgewicht zwischen einem schlanken, effizienten Index und einer robusten, nuancierten Suchfähigkeit, die speziell auf den deutschen Sprachkontext zugeschnitten ist.

**Jetzt können Sie den "german_improved" Analyzer ausprobieren:**

```json
GET german_index/_analyze?pretty
{
  "analyzer": "german_improved", 
  "text": ["Lebensversicherungsgesellschaft."]
}
```

**Sie können einige Dokumente zur Suche hinzufügen und weiter mit dem neuen Index experimentieren.**

```json
PUT /german_index/_doc/1
{
  "paragraph": "Eine Alsterrundfahrt bietet eine einzigartige Gelegenheit, die idyllische Landschaft und die städtische Schönheit Hamburgs vom Wasser aus zu erleben."
} 
```

Nun werden Sie feststellen, dass die Suche dieses Dokument abruft, wenn Sie das Synonym "**Reise**" (bedeutet Reise, Fahrt, Trip) für einen Teil des Wortes "**Alsterrundfahrt**" (Alster Fluss Rundfahrt) verwenden, speziell für "**Rundfahrt**" (Rundfahrt).

```json
GET german_index/_search
{
  "query": {
    "match": {
      "paragraph": "Reise"
    }
  }
}
```

## Fazit

Die Verbesserungen in Amazon OpenSearch mit benutzerdefinierten Textanalyzern zeigen seine Anpassungsfähigkeit und Effizienz bei der Bewältigung deutscher Suchherausforderungen. Dieser Ansatz, der ein nuanciertes Verständnis von Sprachfeinheiten zeigt, ist für Unternehmen und Entwickler, die sich mit Volltextsuche befassen, von unschätzbarem Wert und verbessert sowohl die Benutzererfahrung als auch die Suchrelevanz.

Für diejenigen, die Alternativen suchen, kann das Erforschen von semantischen Suchtechniken und hybriden Ansätzen von Vorteil sein. Die semantische Suche dringt tiefer in das Verständnis des Kontexts und der Bedeutung hinter Benutzeranfragen ein und bietet eine anspruchsvollere Ebene der Suchgenauigkeit. Ein hybrider Ansatz, der die traditionelle Stichwortsuche mit semantischen Fähigkeiten kombiniert, kann die Ergebnisse weiter verfeinern, insbesondere in komplexen Suchszenarien.

Wir ermutigen Unternehmen und Entwickler, diese Verbesserungen in ihren Amazon OpenSearch-Bereitstellungen zu erforschen, um die verbesserten Suchfähigkeiten und betrieblichen Effizienzen aus erster Hand zu erleben.

Wir bei tecRacer können Ihnen bei der Optimierung des Amazon OpenSearch Service helfen. Als [Amazon OpenSearch Service Delivery Partner](https://www.tecracer.com/de/consulting/amazon-opensearch-service/) bieten wir spezialisierte Unterstützung im Infrastrukturdesign, Automatisierung Ihres Cluster-Deployments und -Managements, Monitoring und Suchoptimierung.

— [Alexey](https://www.linkedin.com/comm/mynetwork/discovery-see-all?usecase=PEOPLE_FOLLOWS&followMember=vidanov)