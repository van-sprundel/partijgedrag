package openbesluitvorming

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	sectionHeading      = regexp.MustCompile(`^(constaterende|constaterend|overwegende|overwegend|concluderende|van mening|is van mening|spreekt|verzoekt|vraagt|draagt|roept|geeft (?:het|de) \pL+(?: \pL+)? opdracht)\b`)
	midLineHeading      = regexp.MustCompile(`[;.,]\s+(Constaterende|Overwegende|Spreekt uit|Verzoekt|Vraagt|Draagt|Roept|Geeft (?:het|de) \pL+(?: \pL+)? opdracht)\b`)
	trailingConjunction = regexp.MustCompile(`;\s*en(?:\s+dat)?$`)
	endOfBody           = regexp.MustCompile(`^(en gaat over|en gaat tot de orde|gaat over tot de orde|ingediend door|indiener|indieners|namens|ondertekend|ondertekening|toelichting)\b`)
	closingFormula      = regexp.MustCompile(`(?i)(?:(?:^|\s)\S{1,3}\s+)?\b(?:g?aat over tot de|over tot de orde|tot de orde van de dag)\b`)
	footnote            = regexp.MustCompile(`^\d{1,2}\s*(?:\[|https?:|www\.)`)
	markdownLink        = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	listMarker          = regexp.MustCompile(`^(?:[-•–·*▪]\s*|\d{1,2}[.)]\s+|[a-z][.)]\s+)`)
	pageFooter          = regexp.MustCompile(`\b(pagina|page)\s+\d+(\s+(van|of)\s*\d+)?$`)
	markdownNoise       = strings.NewReplacer("**", "", "__", "", "`", "", "<br>", " ")
	spaces              = regexp.MustCompile(`\s+`)
)

type section struct {
	heading string
	inline  bool
	items   []string
}

func ExtractBulletPoints(markdown string) []string {
	var sections []*section
	var current *section
	itemOpen := false
	inPageBreak := false

	properNouns := midSentenceCapitals(markdown)

	for _, line := range splitLines(markdown) {
		if strings.HasPrefix(line, "|") {
			continue
		}

		lower := strings.ToLower(line)
		if pageFooter.MatchString(lower) || footnote.MatchString(lower) {
			inPageBreak = true
			continue
		}
		if current != nil && !endOfBody.MatchString(lower) {
			if match := closingFormula.FindStringIndex(line); match != nil {
				if before := strings.TrimSpace(line[:match[0]]); before != "" && !inPageBreak {
					appendText(current, &itemOpen, before)
				}
				break
			}
		}
		if endOfBody.MatchString(lower) {
			if current != nil {
				break
			}
			continue
		}

		if sectionHeading.MatchString(lower) {
			inPageBreak = false
			heading, rest, hasColon := strings.Cut(line, ":")
			current = &section{heading: strings.TrimSpace(heading)}
			sections = append(sections, current)
			itemOpen = false
			if !hasColon && isBareHeading(line) {
				continue
			}
			if !hasColon {
				current.heading = ""
				current.inline = true
				current.items = append(current.items, line)
				itemOpen = true
				continue
			}
			if rest = strings.TrimSpace(rest); rest != "" {
				current.items = append(current.items, stripListMarker(rest))
				itemOpen = true
			}
			continue
		}

		if current == nil {
			continue
		}

		if marker := listMarker.FindString(lower); marker != "" && !current.inline {
			text := strings.TrimSpace(line[len(marker):])
			if text == "" && inPageBreak {
				continue
			}
			inPageBreak = false
			current.items = append(current.items, text)
			itemOpen = true
			continue
		}
		if inPageBreak {
			continue
		}

		appendText(current, &itemOpen, line)
	}

	var bullets []string
	for _, section := range sections {
		for _, item := range section.items {
			item = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(item), ","))
			item = trailingConjunction.ReplaceAllString(item, ";")
			item = markdownLink.ReplaceAllString(item, "$1")
			if item == "" {
				continue
			}
			bullet := item
			if section.heading != "" {
				bullet = section.heading + " " + lowerFirst(item, properNouns)
			}
			bullets = append(bullets, bullet)
		}
	}
	return bullets
}

func appendText(current *section, itemOpen *bool, line string) {
	last := len(current.items) - 1
	if !*itemOpen || last < 0 || (!current.inline && startsNewClause(current.items[last], line)) {
		current.items = append(current.items, line)
		*itemOpen = true
		return
	}
	current.items[last] = joinWrapped(current.items[last], line)
}

func isBareHeading(line string) bool {
	if strings.HasSuffix(line, ";") || strings.HasSuffix(line, ".") || strings.HasSuffix(line, ",") {
		return false
	}
	return len(strings.Fields(line)) <= 8
}

func cleanLine(line string) string {
	line = markdownNoise.Replace(line)
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, "#")
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "_")
	line = strings.ReplaceAll(line, " _", " ")
	line = strings.ReplaceAll(line, "_ ", " ")
	line = strings.ReplaceAll(line, " ", " ")
	return spaces.ReplaceAllString(strings.TrimSpace(line), " ")
}

func stripListMarker(text string) string {
	if marker := listMarker.FindString(strings.ToLower(text)); marker != "" {
		return strings.TrimSpace(text[len(marker):])
	}
	return text
}

func joinWrapped(item string, next string) string {
	if strings.HasSuffix(item, "-") {
		first, _ := utf8.DecodeRuneInString(next)
		continues := strings.HasPrefix(next, "en ") || strings.HasPrefix(next, "of ")
		if unicode.IsLower(first) && !continues {
			return strings.TrimSuffix(item, "-") + next
		}
	}
	return item + " " + next
}

func lowerFirst(text string, properNouns map[string]bool) string {
	first, size := utf8.DecodeRuneInString(text)
	if !unicode.IsUpper(first) {
		return text
	}
	second, _ := utf8.DecodeRuneInString(text[size:])
	if !unicode.IsLower(second) {
		return text
	}
	if properNouns[firstWord(text)] {
		return text
	}
	return string(unicode.ToLower(first)) + text[size:]
}

func splitLines(markdown string) []string {
	var lines []string
	for _, rawLine := range strings.Split(markdown, "\n") {
		line := cleanLine(rawLine)
		for line != "" {
			match := midLineHeading.FindStringSubmatchIndex(line)
			if match == nil || strings.HasPrefix(line, "|") {
				lines = append(lines, line)
				break
			}
			lines = append(lines, strings.TrimSpace(line[:match[2]]))
			line = strings.TrimSpace(line[match[2]:])
		}
	}
	return lines
}

func startsNewClause(item string, next string) bool {
	if !strings.HasSuffix(item, ";") && !trailingConjunction.MatchString(item) {
		return false
	}
	first, _ := utf8.DecodeRuneInString(next)
	return unicode.IsUpper(first)
}

func midSentenceCapitals(markdown string) map[string]bool {
	words := map[string]bool{}
	for _, line := range strings.Split(markdown, "\n") {
		line = cleanLine(line)
		if strings.HasPrefix(line, "|") {
			continue
		}
		fields := strings.Fields(line)
		for i := 1; i < len(fields); i++ {
			previous := fields[i-1]
			if strings.ContainsAny(previous[len(previous)-1:], ".:;!?") {
				continue
			}
			first, _ := utf8.DecodeRuneInString(fields[i])
			word := firstWord(fields[i])
			if unicode.IsUpper(first) && !functionWords[strings.ToLower(word)] {
				words[word] = true
			}
		}
	}
	return words
}

var functionWords = map[string]bool{}

func init() {
	for _, word := range strings.Fields(`de het een in op te om voor bij met uit van aan naar over onder tussen door
		dit deze dat die er ook als zo al alle nog niet wel en of maar omdat indien wanneer waar hoe wat
		zij wij hij ze we u hun haar zijn is wordt worden heeft hebben kan kunnen moet moeten zal zullen
		geen veel meer minder steeds daarbij daarom echter tevens bovendien dit inclusief`) {
		functionWords[word] = true
	}
}

func firstWord(text string) string {
	end := strings.IndexFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	if end < 0 {
		return text
	}
	return text[:end]
}
