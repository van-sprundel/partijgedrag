package openbesluitvorming

import (
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Signatures struct {
	names *regexp.Regexp
}

var nationalParties = []string{
	"GroenLinks", "PvdA", "GroenLinks-PvdA", "VVD", "CDA", "D66", "SP", "PvdD",
	"Partij voor de Dieren", "ChristenUnie", "SGP", "Volt", "50PLUS", "JA21",
	"BBB", "PVV", "Forum voor Democratie",
}

var (
	groupPrefix = regexp.MustCompile(`(?i)^(?:fractie|lijst|raadsleden|groep)\s+`)
	camelCase   = regexp.MustCompile(`(\p{Ll})(\p{Lu})`)
	askPreamble = regexp.MustCompile(`(?i)^(?:verzoekt|vraagt|draagt|roept|geeft|spreekt uit|constaterende|overwegende)(?:\s+(?:het|de)\s+\pL+)?(?:\s+(?:op|opdracht))?(?:\s+om)?\b`)
)

var stampWords = map[string]bool{"voor": true, "tegen": true}

func NewSignatures(partyNames []string) *Signatures {
	seen := map[string]bool{}
	var patterns []string
	for _, name := range append(slices.Clone(partyNames), nationalParties...) {
		name = strings.TrimSpace(groupPrefix.ReplaceAllString(strings.TrimSpace(name), ""))
		if utf8.RuneCountInString(name) < 2 || seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true
		patterns = append(patterns, namePattern(name))
	}
	sort.Slice(patterns, func(i, j int) bool { return len(patterns[i]) > len(patterns[j]) })
	return &Signatures{names: regexp.MustCompile(strings.Join(patterns, "|"))}
}

func namePattern(name string) string {
	name = camelCase.ReplaceAllString(name, "$1 $2")
	first, size := utf8.DecodeRuneInString(name)
	head := regexp.QuoteMeta(string(first))
	if lower, upper := unicode.ToLower(first), unicode.ToUpper(first); lower != upper {
		head = "[" + string(lower) + string(upper) + "]"
	}
	parts := strings.Fields(name[size:])
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	rest := strings.Join(parts, `\s*`)
	if strings.HasPrefix(name[size:], " ") {
		rest = `\s*` + rest
	}
	return head + rest
}

func (s *Signatures) Strip(bullets []string) []string {
	out := slices.Clone(bullets)
	for i := len(out) - 1; i >= 0; i-- {
		stripped, cut := s.stripBullet(out[i])
		if !cut {
			break
		}
		if !readable(stripped) {
			out = slices.Delete(out, i, i+1)
			continue
		}
		out[i] = stripped
		break
	}
	return out
}

type span struct{ start, end int }

func (s *Signatures) stripBullet(bullet string) (string, bool) {
	names := s.matchNames(bullet)
	if len(names) == 0 {
		return bullet, false
	}
	tokens := tokenize(bullet)
	for _, token := range tokens[1:] {
		if s.junkTail(bullet, token.start, tokens, names) {
			return strings.TrimSpace(bullet[:token.start]), true
		}
	}
	return bullet, false
}

func (s *Signatures) junkTail(bullet string, from int, tokens []span, names []span) bool {
	distinct := map[string]bool{}
	startsWithName := false
	for _, name := range names {
		if name.start < from {
			continue
		}
		if name.start == from {
			startsWithName = true
		}
		distinct[nameKey(bullet[name.start:name.end])] = true
	}
	if len(distinct) == 0 {
		return false
	}

	shortLowercase := false
	for _, token := range tokens {
		if token.start < from || overlaps(token, names) {
			continue
		}
		text := bullet[token.start:token.end]
		if strings.Contains(text, "://") || strings.HasPrefix(text, "www.") {
			break
		}
		word := strings.TrimFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		if !isLowercaseWord(word) {
			continue
		}
		length := utf8.RuneCountInString(word)
		if length >= 4 && !stampWords[word] {
			return false
		}
		if length >= 2 {
			shortLowercase = true
		}
	}

	if endsSentence(bullet[:from]) || onlyPreamble(bullet[:from]) {
		return true
	}
	return startsWithName && len(distinct) >= 2 && !shortLowercase
}

func (s *Signatures) matchNames(bullet string) []span {
	var names []span
	for _, match := range s.names.FindAllStringIndex(bullet, -1) {
		before, _ := utf8.DecodeLastRuneInString(bullet[:match[0]])
		after, _ := utf8.DecodeRuneInString(bullet[match[1]:])
		if match[0] > 0 && isWordRune(before) || match[1] < len(bullet) && isWordRune(after) {
			continue
		}
		names = append(names, span{match[0], match[1]})
	}
	return names
}

func readable(bullet string) bool {
	rest := askPreamble.ReplaceAllString(bullet, "")
	words := 0
	for _, field := range strings.Fields(rest) {
		word := strings.TrimFunc(field, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		if utf8.RuneCountInString(word) >= 2 && isLowercaseWord(word) {
			words++
		}
	}
	return words >= 3
}

func tokenize(text string) []span {
	var tokens []span
	start := -1
	for i, r := range text {
		if unicode.IsSpace(r) {
			if start >= 0 {
				tokens = append(tokens, span{start, i})
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tokens = append(tokens, span{start, len(text)})
	}
	return tokens
}

func endsSentence(text string) bool {
	last, _ := utf8.DecodeLastRuneInString(strings.TrimRightFunc(text, unicode.IsSpace))
	return last == '.' || last == ';' || last == ':'
}

func onlyPreamble(text string) bool {
	text = strings.TrimSpace(text)
	match := askPreamble.FindStringIndex(text)
	return match != nil && match[1] == len(text)
}

func overlaps(token span, names []span) bool {
	for _, name := range names {
		if token.start < name.end && name.start < token.end {
			return true
		}
	}
	return false
}

func isLowercaseWord(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) || !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func nameKey(name string) string {
	return strings.Map(func(r rune) rune {
		if isWordRune(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, name)
}
