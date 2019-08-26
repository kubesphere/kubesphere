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

package parser

import (
	"errors"
	"fmt"

	"github.com/projectcalico/libcalico-go/lib/selector/tokenizer"
	log "github.com/sirupsen/logrus"
)

const parserDebug = false

// Parse parses a string representation of a selector expression into a Selector.
func Parse(selector string) (sel Selector, err error) {
	log.Debugf("Parsing %#v", selector)
	tokens, err := tokenizer.Tokenize(selector)
	if err != nil {
		return
	}
	if tokens[0].Kind == tokenizer.TokEOF {
		return &selectorRoot{root: &AllNode{}}, nil
	}
	log.Debugf("Tokens %v", tokens)
	// The "||" operator has the lowest precedence so we start with that.
	node, remTokens, err := parseOrExpression(tokens)
	if err != nil {
		return
	}
	if len(remTokens) != 1 {
		err = errors.New(fmt.Sprint("unexpected content at end of selector ", remTokens))
		return
	}
	sel = &selectorRoot{root: node}
	return
}

// parseOrExpression parses a one or more "&&" terms, separated by "||" operators.
func parseOrExpression(tokens []tokenizer.Token) (sel node, remTokens []tokenizer.Token, err error) {
	if parserDebug {
		log.Debugf("Parsing ||s from %v", tokens)
	}
	// Look for the first expression.
	andNodes := make([]node, 0)
	sel, remTokens, err = parseAndExpression(tokens)
	if err != nil {
		return
	}
	andNodes = append(andNodes, sel)

	// Then loop looking for "||" followed by an <expression>
	for {
		switch remTokens[0].Kind {
		case tokenizer.TokOr:
			remTokens = remTokens[1:]
			sel, remTokens, err = parseAndExpression(remTokens)
			if err != nil {
				return
			}
			andNodes = append(andNodes, sel)
		default:
			if len(andNodes) == 1 {
				sel = andNodes[0]
			} else {
				sel = &OrNode{andNodes}
			}
			return
		}
	}
}

// parseAndExpression parses a one or more operations, separated by "&&" operators.
func parseAndExpression(tokens []tokenizer.Token) (sel node, remTokens []tokenizer.Token, err error) {
	if parserDebug {
		log.Debugf("Parsing &&s from %v", tokens)
	}
	// Look for the first operation.
	opNodes := make([]node, 0)
	sel, remTokens, err = parseOperation(tokens)
	if err != nil {
		return
	}
	opNodes = append(opNodes, sel)

	// Then loop looking for "&&" followed by another operation.
	for {
		switch remTokens[0].Kind {
		case tokenizer.TokAnd:
			remTokens = remTokens[1:]
			sel, remTokens, err = parseOperation(remTokens)
			if err != nil {
				return
			}
			opNodes = append(opNodes, sel)
		default:
			if len(opNodes) == 1 {
				sel = opNodes[0]
			} else {
				sel = &AndNode{opNodes}
			}
			return
		}
	}
}

// parseOperations parses a single, possibly negated operation (i.e. ==, !=, has()).
// It also handles calling parseOrExpression recursively for parenthesized expressions.
func parseOperation(tokens []tokenizer.Token) (sel node, remTokens []tokenizer.Token, err error) {
	if parserDebug {
		log.Debugf("Parsing op from %v", tokens)
	}
	if len(tokens) == 0 {
		err = errors.New("Unexpected end of string looking for op")
		return
	}

	// First, collapse any leading "!" operators to a single boolean.
	negated := false
	for {
		if tokens[0].Kind == tokenizer.TokNot {
			negated = !negated
			tokens = tokens[1:]
		} else {
			break
		}
	}

	// Then, look for the various types of operator.
	switch tokens[0].Kind {
	case tokenizer.TokHas:
		sel = &HasNode{tokens[0].Value.(string)}
		remTokens = tokens[1:]
	case tokenizer.TokAll:
		sel = &AllNode{}
		remTokens = tokens[1:]
	case tokenizer.TokLabel:
		// should have an operator and a literal.
		if len(tokens) < 3 {
			err = errors.New(fmt.Sprint("Unexpected end of string in middle of op", tokens))
			return
		}
		switch tokens[1].Kind {
		case tokenizer.TokEq:
			if tokens[2].Kind == tokenizer.TokStringLiteral {
				sel = &LabelEqValueNode{tokens[0].Value.(string), tokens[2].Value.(string)}
				remTokens = tokens[3:]
			} else {
				err = errors.New("Expected string")
			}
		case tokenizer.TokNe:
			if tokens[2].Kind == tokenizer.TokStringLiteral {
				sel = &LabelNeValueNode{tokens[0].Value.(string), tokens[2].Value.(string)}
				remTokens = tokens[3:]
			} else {
				err = errors.New("Expected string")
			}
		case tokenizer.TokContains:
			if tokens[2].Kind == tokenizer.TokStringLiteral {
				sel = &LabelContainsValueNode{tokens[0].Value.(string), tokens[2].Value.(string)}
				remTokens = tokens[3:]
			} else {
				err = errors.New("Expected string")
			}
		case tokenizer.TokStartsWith:
			if tokens[2].Kind == tokenizer.TokStringLiteral {
				sel = &LabelStartsWithValueNode{tokens[0].Value.(string), tokens[2].Value.(string)}
				remTokens = tokens[3:]
			} else {
				err = errors.New("Expected string")
			}
		case tokenizer.TokEndsWith:
			if tokens[2].Kind == tokenizer.TokStringLiteral {
				sel = &LabelEndsWithValueNode{tokens[0].Value.(string), tokens[2].Value.(string)}
				remTokens = tokens[3:]
			} else {
				err = errors.New("Expected string")
			}
		case tokenizer.TokIn, tokenizer.TokNotIn:
			if tokens[2].Kind == tokenizer.TokLBrace {
				remTokens = tokens[3:]
				values := []string{}
				for {
					if remTokens[0].Kind == tokenizer.TokStringLiteral {
						value := remTokens[0].Value.(string)
						values = append(values, value)
						remTokens = remTokens[1:]
						if remTokens[0].Kind == tokenizer.TokComma {
							remTokens = remTokens[1:]
						} else {
							break
						}
					} else {
						break
					}
				}
				if remTokens[0].Kind != tokenizer.TokRBrace {
					err = errors.New("Expected }")
				} else {
					// Skip over the }
					remTokens = remTokens[1:]

					labelName := tokens[0].Value.(string)
					set := ConvertToStringSetInPlace(values) // Mutates values.
					if tokens[1].Kind == tokenizer.TokIn {
						sel = &LabelInSetNode{labelName, set}
					} else {
						sel = &LabelNotInSetNode{labelName, set}
					}
				}
			} else {
				err = errors.New("Expected set literal")
			}
		default:
			err = errors.New(fmt.Sprint("Expected == or != not ", tokens[1]))
			return
		}
	case tokenizer.TokLParen:
		// We hit a paren, skip past it, then recurse.
		sel, remTokens, err = parseOrExpression(tokens[1:])
		if err != nil {
			return
		}
		// After parsing the nested expression, there should be
		// a matching paren.
		if len(remTokens) < 1 || remTokens[0].Kind != tokenizer.TokRParen {
			err = errors.New("Expected )")
			return
		}
		remTokens = remTokens[1:]
	default:
		err = errors.New(fmt.Sprint("Unexpected token: ", tokens[0]))
		return
	}
	if negated && err == nil {
		sel = &NotNode{sel}
	}
	return
}
