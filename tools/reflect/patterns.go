// Correction patterns and scoring adapted from claude-reflect
// (https://github.com/BayramAnnakov/claude-reflect), scripts/lib/reflect_utils.py.
//
// MIT License
//
// Copyright (c) 2025 Bayram Annakov
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxCapturePromptLen = 500
	maxWeakPatternLen   = 150
	minShortCorrection  = 80
)

type rule struct {
	re     *regexp.Regexp
	name   string
	strong bool
	conf   float64
}

func rules(defs ...rule) []rule { return defs }

func r(pattern, name string, strong bool, conf float64) rule {
	return rule{regexp.MustCompile(pattern), name, strong, conf}
}

func res(patterns ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		out[i] = regexp.MustCompile("(?i)" + p)
	}
	return out
}

var (
	slashCommand = regexp.MustCompile(`^/[A-Za-z][\w.-]*(?::[\w.-]+)?(?:\s|$)`)
	cjkChar      = regexp.MustCompile(`[\x{3000}-\x{9fff}\x{f900}-\x{faff}\x{ac00}-\x{d7af}]`)
	explicit     = regexp.MustCompile(`(?i)remember:`)

	guardrails = rules(
		r(`(?i)don't (?:add|include|create) .{1,40} unless`, "dont-unless-asked", true, 0.90),
		r(`(?i)only (?:change|modify|edit|touch) what I (?:asked|requested|said)`, "only-what-asked", true, 0.90),
		r(`(?i)stop (?:refactoring|changing|modifying|editing) (?:unrelated|other|surrounding)`, "stop-unrelated", true, 0.90),
		r(`(?i)don't (?:over-engineer|add extra|be too|make unnecessary)`, "dont-over-engineer", true, 0.85),
		r(`(?i)don't (?:refactor|reorganize|restructure) (?:unless|without)`, "dont-refactor-unless", true, 0.85),
		r(`(?i)leave .{1,30} (?:alone|unchanged|as is)`, "leave-alone", true, 0.85),
		r(`(?i)don't (?:add|include) (?:comments|docstrings|type hints|annotations) (?:unless|to code)`, "dont-add-annotations", true, 0.85),
		r(`(?i)(?:minimal|minimum|only necessary) changes`, "minimal-changes", true, 0.80),
	)

	falsePositives = res(
		`[?\x{ff1f}]$`,
		`[\x{55ce}\x{5417}\x{5462}\x{304b}\x{ae4c}]$`,
		`^(please|can you|could you|would you|help me)\b`,
		`(help|fix|check|review|figure out|set up)\s+(this|that|it|the)\b`,
		`(error|failed|could not|cannot|can't|unable to)\s+\w+`,
		`(is|was|are|were)\s+(not|broken|failing)`,
		`^I (need|want|would like)\b`,
		`^(ok|okay|alright)[,.]?\s+(so|now|let)`,
	)

	nonCorrections = res(
		`^no\s+problem`,
		`^no\s+worries`,
		`^no\s+need\b`,
		`^no\s+way\b`,
		`^don't\s+worry`,
		`^don't\s+mind`,
		`^don't\s+bother`,
		`^never\s+mind`,
		`^no\s+idea\b`,
		`^no\s+(?:it|that|this|we|i|you|they)\s+(?:works?|worked|looks?|seems?|sounds?|reads?)\b`,
		`^no\s+(?:it|that|this)\s+(?:'s|is|was)\s+(?:fine|good|ok|okay|right|correct)\b`,
		`^no\s+(?:i|we)\s+(?:think|guess|believe|reckon)\b`,
		`^no\s+(?:you|we|i)\s+(?:can|could|should)\s+go\s+ahead\b`,
		`^no\s+(?:i|we)(?:'m|'re| am| are)?\s+(?:all\s+)?(?:good|done|set|fine)\b`,
		`^no\s+[\w-]+\s+(?:appeared|happened|occurred|showed|showed\s+up|returned|existed|came|come|changed|matched|was|were|has|have|had)\b`,
		`^no\s+(?:rush|hurry|pressure|problem|stress)\b`,
		`^stop\s+worrying`,
	)

	cjkCorrections = rules(
		r(`^いや[、,.\s]|^いや違`, "iya", true, 0),
		r(`^違う[、，,.\s！!。]|^ちがう[、,.\s]`, "chigau", true, 0),
		r(`そうじゃなく[てけ]|そっちじゃなく[てけ]`, "souja-nakute", true, 0),
		r(`間違[いえっ]て`, "machigatte", true, 0),
		r(`じゃなくて.{0,30}にして`, "janakute-nishite", true, 0),
		r(`^やめて[。！!]?\s*$`, "yamete", true, 0),
		r(`^そうじゃない`, "souja-nai", true, 0),
		r(`って言った[のよでじゃ]`, "tte-itta", true, 0),
		r(`^不是[，,. ]`, "bushi", true, 0),
		r(`^错了|^錯了`, "cuole", true, 0),
		r(`不要.{0,20}要`, "buyao-yao", true, 0),
		r(`^아니[,. ]`, "ani", true, 0),
		r(`틀렸`, "teullyeoss", true, 0),
	)

	englishCorrections = rules(
		r(`(?i)^no[,.!:;\x{2014}\x{2013}-]+\s*\S`, "no,", true, 0),
		r(`(?i)^no\s+\S`, "no-bare", true, 0),
		r(`(?i)^don't\b|^do not\b`, "don't", true, 0),
		r(`(?i)^stop\b|^never\b`, "stop/never", true, 0),
		r(`(?i)that's (wrong|incorrect)|that is (wrong|incorrect)`, "that's-wrong", true, 0),
		r(`(?i)^actually[,. ]`, "actually", false, 0),
		r(`(?i)^I meant\b|^I said\b`, "I-meant/said", true, 0),
		r(`(?i)^I told you\b|^I already told\b`, "I-told-you", true, 0),
		r(`(?i)use .{1,30} not\b`, "use-X-not-Y", true, 0),
	)

	skipMessage = regexp.MustCompile(`^<|^\[|^\{|tool_result|tool_use_id|<command-|<task-notification>|<system-reminder>|This session is being continued|^Analysis:|^\*\*|^   -`)
)

func includeMessage(text string) bool {
	return strings.TrimSpace(text) != "" && !skipMessage.MatchString(text)
}

func matchesAny(res []*regexp.Regexp, text string) bool {
	for _, re := range res {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func matchedNames(rs []rule, text string) (names []string, strong bool) {
	for _, rl := range rs {
		if rl.re.MatchString(text) {
			names = append(names, rl.name)
			strong = strong || rl.strong
		}
	}
	return names, strong
}

func lengthAdjusted(conf float64, n int) float64 {
	switch {
	case n < minShortCorrection:
		return min(0.90, conf+0.10)
	case n > 300:
		return max(0.50, conf-0.15)
	case n > 150:
		return max(0.55, conf-0.10)
	}
	return conf
}

func detectCorrection(text string) (patterns []string, conf float64) {
	text = strings.TrimSpace(text)
	if slashCommand.MatchString(text) {
		return nil, 0
	}
	short := 4
	if cjkChar.MatchString(text) {
		short = 2
	}
	if utf8.RuneCountInString(text) <= short {
		return nil, 0
	}
	if explicit.MatchString(text) {
		return []string{"remember:"}, 0.90
	}
	for _, g := range guardrails {
		if g.re.MatchString(text) {
			return []string{g.name}, g.conf
		}
	}
	if matchesAny(falsePositives, text) || matchesAny(nonCorrections, text) {
		return nil, 0
	}

	n := len(text)
	if names, strong := matchedNames(cjkCorrections, text); len(names) > 0 {
		conf = 0.60
		if strong {
			conf = 0.75
		}
		return names, lengthAdjusted(conf, n)
	}

	var names []string
	var strong, toldYou bool
	for _, rl := range englishCorrections {
		if !rl.re.MatchString(text) || (!rl.strong && n > maxWeakPatternLen) {
			continue
		}
		names = append(names, rl.name)
		strong = strong || rl.strong
		toldYou = toldYou || rl.name == "I-told-you"
	}
	switch {
	case len(names) == 0:
		return nil, 0
	case toldYou, len(names) >= 3:
		conf = 0.85
	case len(names) == 2:
		conf = 0.75
	case strong:
		conf = 0.70
	default:
		conf = 0.55
	}
	return names, lengthAdjusted(conf, n)
}
