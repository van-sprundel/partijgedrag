package openbesluitvorming

import (
	"strings"
	"unicode"
)

var fillerWords = map[string]bool{
	"fractie":              true,
	"raadsfractie":         true,
	"gemeenteraadsfractie": true,
	"raadsleden":           true,
	"raadslid":             true,
	"gemeenteraadsleden":   true,
	"gemeenteraadslid":     true,
	"commissieleden":       true,
	"burgerraadsleden":     true,
	"burgercommissieleden": true,
	"steunfractie":         true,
}

var nationalAliases = map[string]string{
	"partijvoordevrijheid":                "pvv",
	"partijvoordedieren":                  "pvdd",
	"partijvandearbeid":                   "pvda",
	"christendemocratischappel":           "cda",
	"cu":                                  "christenunie",
	"gl":                                  "groenlinks",
	"glpvda":                              "groenlinkspvda",
	"pvdagroenlinks":                      "groenlinkspvda",
	"pvdagl":                              "groenlinkspvda",
	"fvd":                                 "forumvoordemocratie",
	"socialistischepartij":                "sp",
	"staatkundiggereformeerdepartij":      "sgp",
	"volkspartijvoorvrijheidendemocratie": "vvd",
	"boerburgerbeweging":                  "bbb",
	"50plus":                              "50plus",
	"democraten66":                        "d66",
	"cusgp":                               "christenuniesgp",
	"christenuniesgp":                     "christenuniesgp",
}

func PartyKey(groupName string, municipality string) string {
	words := partyWords(groupName)
	if len(words) == 0 {
		return ""
	}

	municipalityWords := partyWords(municipality)
	if len(municipalityWords) > 0 && len(words) > len(municipalityWords) {
		tail := words[len(words)-len(municipalityWords):]
		if strings.Join(tail, " ") == strings.Join(municipalityWords, " ") {
			words = words[:len(words)-len(municipalityWords)]
		}
	}

	key := strings.Join(words, "")
	if alias, ok := nationalAliases[key]; ok {
		key = alias
	}
	return key
}

func partyWords(name string) []string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "&", " en ")
	name = strings.ReplaceAll(name, "+", "plus")
	name = strings.ReplaceAll(name, "-fractie", " ")
	name = strings.ReplaceAll(name, "fractie-", " ")

	fields := strings.FieldsFunc(name, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == '(' || r == ')' || r == '/' || r == ':'
	})

	words := make([]string, 0, len(fields))
	for _, field := range fields {
		compact := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, field)
		if compact == "" || fillerWords[compact] {
			continue
		}
		words = append(words, compact)
	}
	meaningful := false
	for _, word := range words {
		if word != "de" && word != "van" && word != "het" {
			meaningful = true
			break
		}
	}
	if !meaningful {
		return nil
	}
	return words
}
