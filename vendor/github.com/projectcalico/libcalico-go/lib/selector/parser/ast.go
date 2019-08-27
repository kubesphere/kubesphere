// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.

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
	_ "crypto/sha256" // register hash func
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/hash"
)

// Labels defines the interface of labels that can be used by selector
type Labels interface {
	// Get returns value and presence of the given labelName
	Get(labelName string) (value string, present bool)
}

// MapAsLabels allows you use map as labels
type MapAsLabels map[string]string

// Get returns the value and presence of the given labelName key in the MapAsLabels
func (l MapAsLabels) Get(labelName string) (value string, present bool) {
	value, present = l[labelName]
	return
}

// Selector represents a label selector.
type Selector interface {
	// Evaluate evaluates the selector against the given labels expressed as a concrete map.
	Evaluate(labels map[string]string) bool

	// EvaluateLabels evaluates the selector against the given labels expressed as an interface.
	// This allows for labels that are calculated on the fly.
	EvaluateLabels(labels Labels) bool

	// String returns a string that represents this selector.
	String() string

	// UniqueID returns the unique ID that represents this selector.
	UniqueID() string

	// AcceptVisitor allows an external visitor to modify this selector.
	AcceptVisitor(v Visitor)
}

type Visitor interface {
	Visit(n interface{})
}

// PrefixVisitor implements the Visitor interface to allow prefixing of
// label names within a selector.
type PrefixVisitor struct {
	Prefix string
}

func (v PrefixVisitor) Visit(n interface{}) {
	log.Debugf("PrefixVisitor visiting node %#v", n)
	switch np := n.(type) {
	case *LabelEqValueNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelNeValueNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelContainsValueNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelStartsWithValueNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelEndsWithValueNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *HasNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelInSetNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	case *LabelNotInSetNode:
		np.LabelName = fmt.Sprintf("%s%s", v.Prefix, np.LabelName)
	default:
		log.Debug("Node is a no-op")
	}
}

type selectorRoot struct {
	root         node
	cachedString *string
	cachedHash   *string
}

func (sel *selectorRoot) Evaluate(labels map[string]string) bool {
	return sel.EvaluateLabels(MapAsLabels(labels))
}

func (sel *selectorRoot) EvaluateLabels(labels Labels) bool {
	return sel.root.Evaluate(labels)
}

func (sel *selectorRoot) AcceptVisitor(v Visitor) {
	sel.root.AcceptVisitor(v)
}

func (sel *selectorRoot) String() string {
	if sel.cachedString == nil {
		fragments := sel.root.collectFragments([]string{})
		joined := strings.Join(fragments, "")
		sel.cachedString = &joined
	}
	return *sel.cachedString
}

func (sel *selectorRoot) UniqueID() string {
	if sel.cachedHash == nil {
		hash := hash.MakeUniqueID("s", sel.String())
		sel.cachedHash = &hash
	}
	return *sel.cachedHash
}

var _ Selector = (*selectorRoot)(nil)

type node interface {
	Evaluate(labels Labels) bool
	AcceptVisitor(v Visitor)
	collectFragments(fragments []string) []string
}

type LabelEqValueNode struct {
	LabelName string
	Value     string
}

func (node *LabelEqValueNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return val == node.Value
	}
	return false
}

func (node *LabelEqValueNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelEqValueNode) collectFragments(fragments []string) []string {
	return appendLabelOpAndQuotedString(fragments, node.LabelName, " == ", node.Value)
}

type LabelContainsValueNode struct {
	LabelName string
	Value     string
}

func (node *LabelContainsValueNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return strings.Contains(val, node.Value)
	}
	return false
}

func (node *LabelContainsValueNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelContainsValueNode) collectFragments(fragments []string) []string {
	return appendLabelOpAndQuotedString(fragments, node.LabelName, " contains ", node.Value)
}

type LabelStartsWithValueNode struct {
	LabelName string
	Value     string
}

func (node *LabelStartsWithValueNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return strings.HasPrefix(val, node.Value)
	}
	return false
}

func (node *LabelStartsWithValueNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelStartsWithValueNode) collectFragments(fragments []string) []string {
	return appendLabelOpAndQuotedString(fragments, node.LabelName, " starts with ", node.Value)
}

type LabelEndsWithValueNode struct {
	LabelName string
	Value     string
}

func (node *LabelEndsWithValueNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return strings.HasSuffix(val, node.Value)
	}
	return false
}

func (node *LabelEndsWithValueNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelEndsWithValueNode) collectFragments(fragments []string) []string {
	return appendLabelOpAndQuotedString(fragments, node.LabelName, " ends with ", node.Value)
}

type LabelInSetNode struct {
	LabelName string
	Value     StringSet
}

func (node *LabelInSetNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return node.Value.Contains(val)
	}
	return false
}

func (node *LabelInSetNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelInSetNode) collectFragments(fragments []string) []string {
	return collectInOpFragments(fragments, node.LabelName, "in", node.Value)
}

type LabelNotInSetNode struct {
	LabelName string
	Value     StringSet
}

func (node *LabelNotInSetNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelNotInSetNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return !node.Value.Contains(val)
	}
	return true
}

func (node *LabelNotInSetNode) collectFragments(fragments []string) []string {
	return collectInOpFragments(fragments, node.LabelName, "not in", node.Value)
}

// collectInOpFragments is a shared implementation of collectFragments
// for the 'in' and 'not in' operators.
func collectInOpFragments(fragments []string, labelName, op string, values StringSet) []string {
	var quote string
	fragments = append(fragments, labelName, " ", op, " {")
	first := true
	for _, s := range values {
		if strings.Contains(s, `"`) {
			quote = `'`
		} else {
			quote = `"`
		}
		if !first {
			fragments = append(fragments, ", ")
		} else {
			first = false
		}
		fragments = append(fragments, quote, s, quote)
	}
	fragments = append(fragments, "}")
	return fragments
}

type LabelNeValueNode struct {
	LabelName string
	Value     string
}

func (node *LabelNeValueNode) Evaluate(labels Labels) bool {
	val, ok := labels.Get(node.LabelName)
	if ok {
		return val != node.Value
	}
	return true
}

func (node *LabelNeValueNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *LabelNeValueNode) collectFragments(fragments []string) []string {
	return appendLabelOpAndQuotedString(fragments, node.LabelName, " != ", node.Value)
}

type HasNode struct {
	LabelName string
}

func (node *HasNode) Evaluate(labels Labels) bool {
	_, ok := labels.Get(node.LabelName)
	if ok {
		return true
	}
	return false
}

func (node *HasNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *HasNode) collectFragments(fragments []string) []string {
	return append(fragments, "has(", node.LabelName, ")")
}

type NotNode struct {
	Operand node
}

func (node *NotNode) Evaluate(labels Labels) bool {
	return !node.Operand.Evaluate(labels)
}

func (node *NotNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
	node.Operand.AcceptVisitor(v)
}

func (node *NotNode) collectFragments(fragments []string) []string {
	fragments = append(fragments, "!")
	return node.Operand.collectFragments(fragments)
}

type AndNode struct {
	Operands []node
}

func (node *AndNode) Evaluate(labels Labels) bool {
	for _, operand := range node.Operands {
		if !operand.Evaluate(labels) {
			return false
		}
	}
	return true
}

func (node *AndNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
	for _, op := range node.Operands {
		op.AcceptVisitor(v)
	}
}

func (node *AndNode) collectFragments(fragments []string) []string {
	fragments = append(fragments, "(")
	fragments = node.Operands[0].collectFragments(fragments)
	for _, op := range node.Operands[1:] {
		fragments = append(fragments, " && ")
		fragments = op.collectFragments(fragments)
	}
	fragments = append(fragments, ")")
	return fragments
}

type OrNode struct {
	Operands []node
}

func (node *OrNode) Evaluate(labels Labels) bool {
	for _, operand := range node.Operands {
		if operand.Evaluate(labels) {
			return true
		}
	}
	return false
}

func (node *OrNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
	for _, op := range node.Operands {
		op.AcceptVisitor(v)
	}
}

func (node *OrNode) collectFragments(fragments []string) []string {
	fragments = append(fragments, "(")
	fragments = node.Operands[0].collectFragments(fragments)
	for _, op := range node.Operands[1:] {
		fragments = append(fragments, " || ")
		fragments = op.collectFragments(fragments)
	}
	fragments = append(fragments, ")")
	return fragments
}

type AllNode struct {
}

func (node *AllNode) Evaluate(labels Labels) bool {
	return true
}

func (node *AllNode) AcceptVisitor(v Visitor) {
	v.Visit(node)
}

func (node *AllNode) collectFragments(fragments []string) []string {
	return append(fragments, "all()")
}

func appendLabelOpAndQuotedString(fragments []string, label, op, s string) []string {
	var quote string
	if strings.Contains(s, `"`) {
		quote = `'`
	} else {
		quote = `"`
	}
	return append(fragments, label, op, quote, s, quote)
}
