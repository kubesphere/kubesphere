/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gengo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/scanner"
)

// ExtractCommentTags parses comments for lines of the form:
//
//	'marker' + "key=value".
//
// Values are optional; "" is the default.  A tag can be specified more than
// one time and all values are returned.  If the resulting map has an entry for
// a key, the value (a slice) is guaranteed to have at least 1 element.
//
// Example: if you pass "+" for 'marker', and the following lines are in
// the comments:
//
//	+foo=value1
//	+bar
//	+foo=value2
//	+baz="qux"
//
// Then this function will return:
//
//	map[string][]string{"foo":{"value1, "value2"}, "bar": {""}, "baz": {`"qux"`}}
//
// Deprecated: Use ExtractFunctionStyleCommentTags.
func ExtractCommentTags(marker string, lines []string) map[string][]string {
	out := map[string][]string{}
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			continue
		}
		if !strings.HasPrefix(line, marker) {
			continue
		}
		kv := strings.SplitN(line[len(marker):], "=", 2)
		if len(kv) == 2 {
			out[kv[0]] = append(out[kv[0]], kv[1])
		} else if len(kv) == 1 {
			out[kv[0]] = append(out[kv[0]], "")
		}
	}
	return out
}

// ExtractSingleBoolCommentTag parses comments for lines of the form:
//
//	'marker' + "key=value1"
//
// If the tag is not found, the default value is returned.  Values are asserted
// to be boolean ("true" or "false"), and any other value will cause an error
// to be returned.  If the key has multiple values, the first one will be used.
func ExtractSingleBoolCommentTag(marker string, key string, defaultVal bool, lines []string) (bool, error) {
	tags, err := ExtractFunctionStyleCommentTags(marker, []string{key}, lines)
	if err != nil {
		return false, err
	}
	values := tags[key]
	if values == nil {
		return defaultVal, nil
	}
	if values[0].Value == "true" {
		return true, nil
	}
	if values[0].Value == "false" {
		return false, nil
	}
	return false, fmt.Errorf("tag value for %q is not boolean: %q", key, values[0])
}

// ExtractFunctionStyleCommentTags parses comments for special metadata tags. The
// marker argument should be unique enough to identify the tags needed, and
// should not be a marker for tags you don't want, or else the caller takes
// responsibility for making that distinction.
//
// The tagNames argument is a list of specific tags being extracted. If this is
// nil or empty, all lines which match the marker are considered.  If this is
// specified, only lines with begin with marker + one of the tags will be
// considered.  This is useful when a common marker is used which may match
// lines which fail this syntax (e.g. which predate this definition).
//
// This function looks for input lines of the following forms:
//   - 'marker' + "key=value"
//   - 'marker' + "key()=value"
//   - 'marker' + "key(arg)=value"
//   - 'marker' + "key(`raw string`)=value"
//   - 'marker' + "key({"k1": "value1"})=value"
//
// The arg is optional.  It may be a Go identifier, a raw string literal
// enclosed in back-ticks, or an object or array represented with a subset of
// JSON syntax. If not specified (either as "key=value" or as
// "key()=value"), the resulting Tag will have an empty Args list.
//
// The value is optional.  If not specified, the resulting Tag will have "" as
// the value.
//
// Tag comment-lines may have a trailing end-of-line comment.
//
// The map returned here is keyed by the Tag's name without args.
//
// A tag can be specified more than one time and all values are returned.  If
// the resulting map has an entry for a key, the value (a slice) is guaranteed
// to have at least 1 element.
//
// Example: if you pass "+" as the marker, and the following lines are in
// the comments:
//
//		+foo=val1  // foo
//		+bar
//		+foo=val2  // also foo
//		+baz="qux"
//		+foo(arg)  // still foo
//	 +buzz({"a": 1, "b": "x"})
//
// Then this function will return:
//
//		map[string][]Tag{
//	 	"foo": []Tag{{
//				Name: "foo",
//				Args: nil,
//				Value: "val1",
//			}, {
//				Name: "foo",
//				Args: nil,
//				Value: "val2",
//			}, {
//				Name: "foo",
//				Args: []string{"arg"},
//				Value: "",
//			}, {
//				Name: "bar",
//				Args: nil,
//				Value: ""
//			}, {
//				Name: "baz",
//				Args: nil,
//				Value: "\"qux\""
//			}, {
//				Name: "buzz",
//				Args: []string{"{\"a\": 1, \"b\": \"x\"}"},
//				Value: ""
//		}}
//
// This function should be preferred instead of ExtractCommentTags.
func ExtractFunctionStyleCommentTags(marker string, tagNames []string, lines []string) (map[string][]Tag, error) {
	// TODO: Both the strings of nested tags and the value of tags might contain //
	//       resulting in a unsound removal of a trailing comment.
	//       This should be fixed by using a grammar to parse the entire tag.
	stripTrailingComment := func(in string) string {
		idx := strings.LastIndex(in, "//")
		if idx == -1 {
			return in
		}
		return strings.TrimSpace(in[:idx])
	}

	out := map[string][]Tag{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, marker) {
			continue
		}
		line = line[len(marker):]
		s := initTagKeyScanner(line)

		name, args, err := s.parseTagKey(tagNames)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		tag := Tag{Name: name, Args: args}
		if s.Scan() == '=' {
			tag.Value = stripTrailingComment(line[s.Offset+1:])
		}
		out[tag.Name] = append(out[tag.Name], tag)
	}
	return out, nil
}

// Tag represents a single comment tag.
type Tag struct {
	// Name is the name of the tag with no arguments.
	Name string
	// Args is a list of optional arguments to the tag.
	Args []string
	// Value is the value of the tag.
	Value string
}

func (t Tag) String() string {
	buf := bytes.Buffer{}
	buf.WriteString(t.Name)
	if len(t.Args) > 0 {
		buf.WriteString("(")
		for i, a := range t.Args {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(a)
		}
		buf.WriteString(")")
	}
	return buf.String()
}

// parseTagKey parses the key part of an extended comment tag, including
// optional arguments. The input is assumed to be the entire text of the
// original input after the marker, up to the '=' or end-of-line.
//
// The tags argument is an optional list of tag names to match. If it is nil or
// empty, all tags match.
//
// At the moment, arguments are very strictly formatted (see parseTagArgs) and
// whitespace is not allowed.
//
// This function returns the key name and arguments, unless tagNames was
// specified and the input did not match, in which case it returns "".
func (s *tagKeyScanner) parseTagKey(tagNames []string) (string, []string, error) {
	if s.Scan() != scanner.Ident {
		return "", nil, fmt.Errorf("expected identifier but got %q", s.TokenText())
	}
	tagName := s.TokenText()
	if len(tagNames) > 0 && !slices.Contains(tagNames, tagName) {
		return "", nil, nil
	}
	if s.PeekSkipSpaces() != '(' {
		return tagName, nil, nil
	}
	s.Scan() // consume the '(' token
	args, err := s.parseTagArgs()
	if err != nil {
		return "", nil, err
	}
	if s.Scan() != ')' {
		return "", nil, s.unexpectedTokenError("')'", s.TokenText())
	}
	return tagName, args, nil
}

// parseTagArgs parses the arguments part of an extended comment tag. The input
// is assumed to be the entire text of the original input after the opening
// '(', and before the trailing ')'.
//
// The argument may be a go style identifier, a quoted string ("..."), or a raw string (`...`).
func (s *tagKeyScanner) parseTagArgs() ([]string, error) {
	if s.PeekSkipSpaces() == ')' {
		return nil, nil
	}
	if s.PeekSkipSpaces() == '{' || s.PeekSkipSpaces() == '[' {
		value, err := s.scanJSONFlavoredValue()
		if err != nil {
			return nil, err
		}
		return []string{value}, nil
	}
	switch s.Scan() {
	case scanner.String, scanner.RawString, scanner.Ident:
		return []string{s.TokenText()}, nil
	default:
		return nil, s.unexpectedTokenError("identifier, quoted string (\"...\") or raw string (`...`)", s.TokenText())
	}
}

// scanJSONFlavoredValue consumes a single token as a JSON value from the scanner and returns the token text.
// A strict subset of JSON is supported, in particular:
// - Big numbers and numbers with exponents are not supported.
// - JSON is expected to be in a single line. Tabs and newlines are not fully supported.
func (s *tagKeyScanner) scanJSONFlavoredValue() (string, error) {
	start, end, err := s.chompJSONFlavoredValue()
	if err != nil {
		return "", err
	}
	value := s.input[start:end]
	var out any
	err = json.Unmarshal([]byte(value), &out) // make sure the JSON parses
	if err != nil {
		return "", err
	}
	return value, nil
}

// chompJSONFlavoredValue consumes valid JSON from the scanner's token stream and returns the start and end positions of the JSON.
func (s *tagKeyScanner) chompJSONFlavoredValue() (int, int, error) {
	switch s.PeekSkipSpaces() {
	case '[':
		return s.chompJSONFlavoredArray()
	case '{':
		return s.chompJSONFlavoredObject()
	}

	t := s.Scan()
	startPos := s.Offset
	switch t {
	case '-', '+':
		t := s.Scan()
		if !(t == scanner.Int || t == scanner.Float) {
			return 0, 0, s.unexpectedTokenError("number", s.TokenText())
		}
		return startPos, s.Offset + len(s.TokenText()), nil
	case scanner.String, scanner.Int, scanner.Float:
		return startPos, s.Offset + len(s.TokenText()), nil
	case scanner.Ident:
		text := s.TokenText()
		if text == "true" || text == "false" || text == "null" {
			return startPos, s.Offset + len(s.TokenText()), nil
		}
	}
	return 0, 0, s.unexpectedTokenError("JSON value", s.TokenText())
}

func (s *tagKeyScanner) chompJSONFlavoredObject() (int, int, error) {
	if s.Scan() != '{' {
		return 0, 0, s.unexpectedTokenError("JSON array", s.TokenText())
	}
	startPos := s.Offset
	if s.PeekSkipSpaces() == '}' {
		s.Scan() // consume }
		return startPos, s.Offset + 1, nil
	}
	_, _, err := s.chompJSONFlavoredObjectEntries()
	if err != nil {
		return 0, 0, err
	}
	if s.Scan() != '}' {
		return 0, 0, s.unexpectedTokenError("}", s.TokenText())
	}
	return startPos, s.Offset + 1, nil
}

func (s *tagKeyScanner) chompJSONFlavoredObjectEntries() (int, int, error) {
	keyStart, _, err := s.chompJSONFlavoredValue()
	if err != nil {
		return 0, 0, err
	}

	if s.Scan() != ':' {
		return 0, 0, s.unexpectedTokenError(":", s.TokenText())
	}

	_, valueEnd, err := s.chompJSONFlavoredValue()
	if err != nil {
		return 0, 0, err
	}

	switch s.PeekSkipSpaces() {
	case ',':
		s.Scan() // Consume ,
		_, entriesEnd, err := s.chompJSONFlavoredObjectEntries()
		if err != nil {
			return 0, 0, err
		}
		return keyStart, entriesEnd, nil
	case '}':
		return keyStart, valueEnd, nil
	default:
		return 0, 0, s.unexpectedTokenError(", or ]", s.TokenText())
	}
}

func (s *tagKeyScanner) chompJSONFlavoredArray() (int, int, error) {
	if s.Scan() != '[' {
		return 0, 0, s.unexpectedTokenError("JSON array", s.TokenText())
	}
	startPos := s.Offset
	if s.PeekSkipSpaces() == ']' {
		s.Scan() // consume ]
		return startPos, s.Offset + 1, nil
	}
	_, _, err := s.chompJSONFlavoredArrayItems()
	if err != nil {
		return 0, 0, err
	}
	if s.Scan() != ']' {
		return 0, 0, s.unexpectedTokenError("]", s.TokenText())
	}
	return startPos, s.Offset + 1, nil
}

func (s *tagKeyScanner) chompJSONFlavoredArrayItems() (int, int, error) {
	valueStart, valueEnd, err := s.chompJSONFlavoredValue()
	if err != nil {
		return 0, 0, err
	}

	switch s.PeekSkipSpaces() {
	case ',':
		s.Scan() // Consume ,
		_, itemsEnd, err := s.chompJSONFlavoredArrayItems()
		if err != nil {
			return 0, 0, err
		}
		return valueStart, itemsEnd, nil
	case ']':
		return valueStart, valueEnd, nil
	default:
		return 0, 0, s.unexpectedTokenError(", or ]", s.TokenText())
	}
}

type tagKeyScanner struct {
	input string
	*scanner.Scanner
	errs []error
}

func initTagKeyScanner(input string) *tagKeyScanner {
	s := tagKeyScanner{input: input, Scanner: &scanner.Scanner{}}
	s.Init(strings.NewReader(input))
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanInts | scanner.ScanFloats

	s.Error = func(scanner *scanner.Scanner, msg string) {
		s.errs = append(s.errs, fmt.Errorf("error parsing '%s' at %v: %s", input, scanner.Position, msg))
	}
	return &s
}

func (s *tagKeyScanner) PeekSkipSpaces() rune {
	ch := s.Peek()
	for ch == ' ' {
		s.Next() // Consume the ' '
		ch = s.Peek()
	}
	return ch
}

func (s *tagKeyScanner) unexpectedTokenError(expected string, token string) error {
	s.Error(s.Scanner, fmt.Sprintf("expected %s but got (%q)", expected, token))
	return errors.Join(s.errs...)
}
