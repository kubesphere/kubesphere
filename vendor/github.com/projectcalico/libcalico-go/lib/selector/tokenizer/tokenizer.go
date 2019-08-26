// Copyright (c) 2016, 2019 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tokenizer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type tokenKind uint8

const (
	TokNone tokenKind = iota
	TokLabel
	TokStringLiteral
	TokLBrace
	TokRBrace
	TokComma
	TokEq
	TokNe
	TokIn
	TokNot
	TokNotIn
	TokContains
	TokStartsWith
	TokEndsWith
	TokAll
	TokHas
	TokLParen
	TokRParen
	TokAnd
	TokOr
	TokEOF
)

const tokenizerDebug = false

var whitespace = " \t"

// Token has a kind and a value
type Token struct {
	Kind  tokenKind
	Value interface{}
}

const (
	// LabelKeyMatcher is the base regex for a valid label key.
	LabelKeyMatcher = `[a-zA-Z0-9_./-]{1,512}`
	hasExpr         = `has\(\s*(` + LabelKeyMatcher + `)\s*\)`
	allExpr         = `all\(\s*\)`
	notInExpr       = `not\s*in\b`
	inExpr          = `in\b`
)

var (
	identifierRegex = regexp.MustCompile("^" + LabelKeyMatcher)
	containsRegex   = regexp.MustCompile(`^contains`)
	startsWithRegex = regexp.MustCompile(`^starts\s*with`)
	endsWithRegex   = regexp.MustCompile(`^ends\s*with`)
	hasRegex        = regexp.MustCompile("^" + hasExpr)
	allRegex        = regexp.MustCompile("^" + allExpr)
	notInRegex      = regexp.MustCompile("^" + notInExpr)
	inRegex         = regexp.MustCompile("^" + inExpr)
)

// Tokenize transforms string to token slice
func Tokenize(input string) (tokens []Token, err error) {
	for {
		if tokenizerDebug {
			log.Debug("Remaining input: ", input)
		}
		startLen := len(input)
		input = strings.TrimLeft(input, whitespace)
		if len(input) == 0 {
			tokens = append(tokens, Token{TokEOF, nil})
			return
		}
		var lastTokKind = TokNone
		if len(tokens) > 0 {
			lastTokKind = tokens[len(tokens)-1].Kind
		}
		switch input[0] {
		case '(':
			tokens = append(tokens, Token{TokLParen, nil})
			input = input[1:]
		case ')':
			tokens = append(tokens, Token{TokRParen, nil})
			input = input[1:]
		case '"':
			input = input[1:]
			index := strings.Index(input, `"`)
			if index == -1 {
				return nil, errors.New("unterminated string")
			}
			value := input[0:index]
			tokens = append(tokens, Token{TokStringLiteral, value})
			input = input[index+1:]
		case '\'':
			input = input[1:]
			index := strings.Index(input, `'`)
			if index == -1 {
				return nil, errors.New("unterminated string")
			}
			value := input[0:index]
			tokens = append(tokens, Token{TokStringLiteral, value})
			input = input[index+1:]
		case '{':
			tokens = append(tokens, Token{TokLBrace, nil})
			input = input[1:]
		case '}':
			tokens = append(tokens, Token{TokRBrace, nil})
			input = input[1:]
		case ',':
			tokens = append(tokens, Token{TokComma, nil})
			input = input[1:]
		case '=':
			if len(input) > 1 && input[1] == '=' {
				tokens = append(tokens, Token{TokEq, nil})
				input = input[2:]
			} else {
				return nil, errors.New("expected ==")
			}
		case '!':
			if len(input) > 1 && input[1] == '=' {
				tokens = append(tokens, Token{TokNe, nil})
				input = input[2:]
			} else {
				tokens = append(tokens, Token{TokNot, nil})
				input = input[1:]
			}
		case '&':
			if len(input) > 1 && input[1] == '&' {
				tokens = append(tokens, Token{TokAnd, nil})
				input = input[2:]
			} else {
				return nil, errors.New("expected &&")
			}
		case '|':
			if len(input) > 1 && input[1] == '|' {
				tokens = append(tokens, Token{TokOr, nil})
				input = input[2:]
			} else {
				return nil, errors.New("expected ||")
			}
		default:
			// Handle less-simple cases with regex matches.  We've already stripped any whitespace.
			if lastTokKind == TokLabel {
				// If we just saw a label, look for a contains/starts with/ends with operator instead of another label.
				if idxs := containsRegex.FindStringIndex(input); idxs != nil {
					// Found "all"
					tokens = append(tokens, Token{TokContains, nil})
					input = input[idxs[1]:]
				} else if idxs := startsWithRegex.FindStringIndex(input); idxs != nil {
					// Found "all"
					tokens = append(tokens, Token{TokStartsWith, nil})
					input = input[idxs[1]:]
				} else if idxs := endsWithRegex.FindStringIndex(input); idxs != nil {
					// Found "all"
					tokens = append(tokens, Token{TokEndsWith, nil})
					input = input[idxs[1]:]
				} else if idxs := notInRegex.FindStringIndex(input); idxs != nil {
					// Found "not in"
					tokens = append(tokens, Token{TokNotIn, nil})
					input = input[idxs[1]:]
				} else if idxs := inRegex.FindStringIndex(input); idxs != nil {
					// Found "in"
					tokens = append(tokens, Token{TokIn, nil})
					input = input[idxs[1]:]
				} else {
					err = fmt.Errorf("unexpected characters after label '%v', was expecting an operator",
						tokens[len(tokens)-1].Value)
					return
				}
			} else if idxs := hasRegex.FindStringSubmatchIndex(input); idxs != nil {
				// Found "has(label)"
				wholeMatchEnd := idxs[1]
				labelNameMatchStart := idxs[2]
				labelNameMatchEnd := idxs[3]
				labelName := input[labelNameMatchStart:labelNameMatchEnd]
				tokens = append(tokens, Token{TokHas, labelName})
				input = input[wholeMatchEnd:]
			} else if idxs := allRegex.FindStringIndex(input); idxs != nil {
				// Found "all"
				tokens = append(tokens, Token{TokAll, nil})
				input = input[idxs[1]:]
			} else if idxs := identifierRegex.FindStringIndex(input); idxs != nil {
				// Found "label"
				endIndex := idxs[1]
				identifier := input[:endIndex]
				log.Debug("Identifier ", identifier)
				tokens = append(tokens, Token{TokLabel, identifier})
				input = input[endIndex:]
			} else {
				err = errors.New("unexpected characters")
				return
			}
		}
		if len(input) >= startLen {
			err = errors.New("infinite loop detected in tokenizer")
			return
		}
	}
}
