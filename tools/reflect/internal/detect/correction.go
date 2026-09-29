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

package detect

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxPromptLen      = 500
	maxWeakPatternLen = 150
	shortCorrection   = 80
)

type rule struct {
	re     *regexp.Regexp
	name   string
	strong bool
}

type guardrail struct {
	re   *regexp.Regexp
	name string
	conf float64
}

func newRule(pattern, name string, strong bool) rule {
	return rule{regexp.MustCompile(pattern), name, strong}
}

func anyOf(patterns ...string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		out[i] = regexp.MustCompile("(?i)" + p)
	}
	return out
}

func guardrailRule(pattern, name string, conf float64) guardrail {
	return guardrail{regexp.MustCompile("(?i)" + pattern), name, conf}
}

var (
	slashCommand = regexp.MustCompile(`^/[A-Za-z][\w.-]*(?::[\w.-]+)?(?:\s|$)`)
	cjkChar      = regexp.MustCompile(`[\x{3000}-\x{9fff}\x{f900}-\x{faff}\x{ac00}-\x{d7af}]`)
	remember     = regexp.MustCompile(`(?i)remember:`)
	skipMessage  = regexp.MustCompile(`^<|^\[|^\{|tool_result|tool_use_id|<command-|<task-notification>|<system-reminder>|This session is being continued|^Analysis:|^\*\*|^   -`)

	guardrails = []guardrail{
		guardrailRule(`don't (?:add|include|create) .{1,40} unless`, "dont-unless-asked", 0.90),
		guardrailRule(`only (?:change|modify|edit|touch) what I (?:asked|requested|said)`, "only-what-asked", 0.90),
		guardrailRule(`stop (?:refactoring|changing|modifying|editing) (?:unrelated|other|surrounding)`, "stop-unrelated", 0.90),
		guardrailRule(`don't (?:over-engineer|add extra|be too|make unnecessary)`, "dont-over-engineer", 0.85),
		guardrailRule(`don't (?:refactor|reorganize|restructure) (?:unless|without)`, "dont-refactor-unless", 0.85),
		guardrailRule(`leave .{1,30} (?:alone|unchanged|as is)`, "leave-alone", 0.85),
		guardrailRule(`don't (?:add|include) (?:comments|docstrings|type hints|annotations) (?:unless|to code)`, "dont-add-annotations", 0.85),
		guardrailRule(`(?:minimal|minimum|only necessary) changes`, "minimal-changes", 0.80),
	}

	notCorrections = anyOf(
		`[?\x{ff1f}]$`,
		`[\x{55ce}\x{5417}\x{5462}\x{304b}\x{ae4c}]$`,
		`^(please|can you|could you|would you|help me)\b`,
		`(help|fix|check|review|figure out|set up)\s+(this|that|it|the)\b`,
		`(error|failed|could not|cannot|can't|unable to)\s+\w+`,
		`(is|was|are|were)\s+(not|broken|failing)`,
		`^I (need|want|would like)\b`,
		`^(ok|okay|alright)[,.]?\s+(so|now|let)`,
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

	cjkCorrections = []rule{
		newRule(`^いや[、,.\s]|^いや違`, "iya", true),
		newRule(`^違う[、，,.\s！!。]|^ちがう[、,.\s]`, "chigau", true),
		newRule(`そうじゃなく[てけ]|そっちじゃなく[てけ]`, "souja-nakute", true),
		newRule(`間違[いえっ]て`, "machigatte", true),
		newRule(`じゃなくて.{0,30}にして`, "janakute-nishite", true),
		newRule(`^やめて[。！!]?\s*$`, "yamete", true),
		newRule(`^そうじゃない`, "souja-nai", true),
		newRule(`って言った[のよでじゃ]`, "tte-itta", true),
		newRule(`^不是[，,. ]`, "bushi", true),
		newRule(`^错了|^錯了`, "cuole", true),
		newRule(`不要.{0,20}要`, "buyao-yao", true),
		newRule(`^아니[,. ]`, "ani", true),
		newRule(`틀렸`, "teullyeoss", true),
	}

	englishCorrections = []rule{
		newRule(`(?i)^no[,.!:;\x{2014}\x{2013}-]+\s*\S`, "no,", true),
		newRule(`(?i)^no\s+\S`, "no-bare", true),
		newRule(`(?i)^don't\b|^do not\b`, "don't", true),
		newRule(`(?i)^stop\b|^never\b`, "stop/never", true),
		newRule(`(?i)that's (wrong|incorrect)|that is (wrong|incorrect)`, "that's-wrong", true),
		newRule(`(?i)^actually[,. ]`, "actually", false),
		newRule(`(?i)^I meant\b|^I said\b`, "I-meant/said", true),
		newRule(`(?i)^I told you\b|^I already told\b`, "I-told-you", true),
		newRule(`(?i)use .{1,30} not\b`, "use-X-not-Y", true),
	}
)

func IncludeMessage(text string) bool {
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

func lengthAdjusted(conf float64, n int) float64 {
	switch {
	case n < shortCorrection:
		return min(0.90, conf+0.10)
	case n > 300:
		return max(0.50, conf-0.15)
	case n > 150:
		return max(0.55, conf-0.10)
	}
	return conf
}

func Correction(text string) (patterns []string, conf float64) {
	text = strings.TrimSpace(text)
	isRemember := remember.MatchString(text)
	switch {
	case slashCommand.MatchString(text):
		return nil, 0
	case len(text) > maxPromptLen && !isRemember:
		return nil, 0
	case utf8.RuneCountInString(text) <= shortThreshold(text):
		return nil, 0
	case isRemember:
		return []string{"remember:"}, 0.90
	}
	for _, g := range guardrails {
		if g.re.MatchString(text) {
			return []string{g.name}, g.conf
		}
	}
	if matchesAny(notCorrections, text) {
		return nil, 0
	}

	n := len(text)
	if names, _ := matched(cjkCorrections, text, n); len(names) > 0 {
		return names, lengthAdjusted(0.75, n)
	}
	names, strong := matched(englishCorrections, text, n)
	switch {
	case len(names) == 0:
		return nil, 0
	case contains(names, "I-told-you"), len(names) >= 3:
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

func shortThreshold(text string) int {
	if cjkChar.MatchString(text) {
		return 2
	}
	return 4
}

func matched(rs []rule, text string, n int) (names []string, strong bool) {
	for _, rl := range rs {
		if !rl.re.MatchString(text) || (!rl.strong && n > maxWeakPatternLen) {
			continue
		}
		names = append(names, rl.name)
		strong = strong || rl.strong
	}
	return names, strong
}

func contains(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}
