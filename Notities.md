## Kamerstukken (een deel hiervan is moties)
> https://gegevensmagazijn.tweedekamer.nl/SyncFeed/2.0/Feed?category=Kamerstukdossier

Entry:
```xml
<entry>
<title>1f031e16-cb3b-45b5-b3c9-a8abd27c913a</title>
<id>
https://gegevensmagazijn.tweedekamer.nl/SyncFeed/2.0/Entiteiten/1f031e16-cb3b-45b5-b3c9-a8abd27c913a
</id>
<content type="application/xml">
	<kamerstukdossier id="1f031e16-cb3b-45b5-b3c9-a8abd27c913a" tk:bijgewerkt="2008-08-26T12:13:04.6270000" tk:verwijderd="false">
		<nummer>30218</nummer> <<<<<
	</kamerstukdossier>
</content>
</entry>
```


## Pages for the site
> Important to understand here is that (incase we want to add a page) it should be easy to make a new visualisation when needed
### Exists
- Stemwijzer wizard
- Stemmen
- Stemuitslag
  - Expand beleidsagenda to see Moties
  - Moties often have multiple begeleidsagenda's
- Moties reference
  - votes per motie
  - original text is click-through-able
- Moties (partijkeuze)
- Partijgelijkenis (Totaal)
- Partijgelijkenis (Per partij en belegeidsagenda)
- Partijfocus (exists but needs cleanup)
  - Technical source is changing
  - Need to read out the moties from the XML file (see example below)
### New features
- Kamerleden/hoofdelijke stemmingen
  - Soms stemmen MPen tegen hun eigen partij.
  - Aggregated hoofdelijke stemmingen
    - Majority rules
    - Tie-breaker: party leader
- Date and/or kabinetsperiode filter (older votes aren't as relevant. keep track of this)
- Onzinmoties/Free beer moties (moties die eigenlijk niks toevoegen. Mag gebruikt worden als hall of shame voor de voorstellende)
- Wetsvoorstellen/amandemeneten (met een focus om ze begrijpbaar te maken. mogelijkheid om de frontpage te alignen hiermee)
- Maybe Front page met motie van de dag (summary voor de luie donder. Niks mis mee)

 `nummer` -> `https://zoek.officielebekendmakingen.nl/kst-[nummer][-n].xml`
  e.g. `https://zoek.officielebekendmakingen.nl/kst-36284-1.xml`

## Terminology rundown
- Wetsvoorstel          - Lange lappen teksten met voorstellen
- Amendementen          - Wetsveranderingsvoorstellen (leuk galgwoordje)
- Moties                - Call to Action (dit boeit ons het meest)

- Begeleidsagenda       - Taxonomie van onderwerpen
- Handelingen           - Ontmoetingsjournaal tk
- Hoofdelijke stemming  - Stemming per persoon ipv per partij (nutteloos, duurtlang)
