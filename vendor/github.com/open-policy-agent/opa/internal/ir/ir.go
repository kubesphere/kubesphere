// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package ir defines an intermediate representation (IR) for Rego.
//
// The IR specifies an imperative execution model for Rego policies similar to a
// query plan in traditional databases.
package ir

import (
	"fmt"
)

type (
	// Policy represents a planned policy query.
	Policy struct {
		Static *Static
		Plan   *Plan
		Funcs  *Funcs
	}

	// Static represents a static data segment that is indexed into by the policy.
	Static struct {
		Strings      []*StringConst
		BuiltinFuncs []*BuiltinFunc
	}

	// BuiltinFunc represents a built-in function that may be required by the
	// policy.
	BuiltinFunc struct {
		Name string
	}

	// Funcs represents a collection of planned functions to include in the
	// policy.
	Funcs struct {
		Funcs []*Func
	}

	// Func represents a named plan (function) that can be invoked. Functions
	// accept one or more parameters and return a value. By convention, the
	// input document and data documents are always passed as the first and
	// second arguments (respectively).
	Func struct {
		Name   string
		Params []Local
		Return Local
		Blocks []*Block // TODO(tsandall): should this be a plan?
	}

	// Plan represents an ordered series of blocks to execute. Plan execution
	// stops when a return statement is reached. Blocks are executed in-order.
	Plan struct {
		Blocks []*Block
	}

	// Block represents an ordered sequence of statements to execute. Blocks are
	// executed until a return statement is encountered, a statement is undefined,
	// or there are no more statements. If all statements are defined but no return
	// statement is encountered, the block is undefined.
	Block struct {
		Stmts []Stmt
	}

	// Stmt represents an operation (e.g., comparison, loop, dot, etc.) to execute.
	Stmt interface {
	}

	// Local represents a plan-scoped variable.
	//
	// TODO(tsandall): should this be int32 for safety?
	Local int

	// Const represents a constant value from the policy.
	Const interface {
		typeMarker()
	}

	// NullConst represents a null value.
	NullConst struct{}

	// BooleanConst represents a boolean value.
	BooleanConst struct {
		Value bool
	}

	// StringConst represents a string value.
	StringConst struct {
		Value string
	}

	// IntConst represents an integer constant.
	IntConst struct {
		Value int64
	}

	// FloatConst represents a floating-point constant.
	FloatConst struct {
		Value float64
	}
)

const (
	// Input is the local variable that refers to the global input document.
	Input Local = iota

	// Data is the local variable that refers to the global data document.
	Data

	// Unused is the free local variable that can be allocated in a plan.
	Unused
)

func (a *Policy) String() string {
	return "Policy"
}

func (a *Static) String() string {
	return fmt.Sprintf("Static (%d strings)", len(a.Strings))
}

func (a *Funcs) String() string {
	return fmt.Sprintf("Funcs (%d funcs)", len(a.Funcs))
}

func (a *Func) String() string {
	return fmt.Sprintf("%v (%d params: %v, %d blocks)", a.Name, len(a.Params), a.Params, len(a.Blocks))
}

func (a *Plan) String() string {
	return fmt.Sprintf("Plan (%d blocks)", len(a.Blocks))
}

func (a *Block) String() string {
	return fmt.Sprintf("Block (%d statements)", len(a.Stmts))
}

func (a *BooleanConst) typeMarker() {}
func (a *NullConst) typeMarker()    {}
func (a *IntConst) typeMarker()     {}
func (a *FloatConst) typeMarker()   {}
func (a *StringConst) typeMarker()  {}

// ReturnLocalStmt represents a return statement that yields a local value.
type ReturnLocalStmt struct {
	Source Local
}

// CallStmt represents a named function call. The result should be stored in the
// result local.
type CallStmt struct {
	Func   string
	Args   []Local
	Result Local
}

// BlockStmt represents a nested block. Nested blocks and break statements can
// be used to short-circuit execution.
type BlockStmt struct {
	Blocks []*Block
}

func (a *BlockStmt) String() string {
	return fmt.Sprintf("BlockStmt (%d blocks)", len(a.Blocks))
}

// BreakStmt represents a jump out of the current block. The index specifies how
// many blocks to jump starting from zero (the current block). Execution will
// continue from the end of the block that is jumped to.
type BreakStmt struct {
	Index uint32
}

// DotStmt represents a lookup operation on a value (e.g., array, object, etc.)
// The source of a DotStmt may be a scalar value in which case the statement
// will be undefined.
type DotStmt struct {
	Source Local
	Key    Local
	Target Local
}

// LenStmt represents a length() operation on a local variable. The
// result is stored in the target local variable.
type LenStmt struct {
	Source Local
	Target Local
}

// ScanStmt represents a linear scan over a composite value. The
// source may be a scalar in which case the block will never execute.
type ScanStmt struct {
	Source Local
	Key    Local
	Value  Local
	Block  *Block
}

// NotStmt represents a negated statement.
type NotStmt struct {
	Block *Block
}

// AssignBooleanStmt represents an assignment of a boolean value to a local variable.
type AssignBooleanStmt struct {
	Value  bool
	Target Local
}

// AssignIntStmt represents an assignment of an integer value to a
// local variable.
type AssignIntStmt struct {
	Value  int64
	Target Local
}

// AssignVarStmt represents an assignment of one local variable to another.
type AssignVarStmt struct {
	Source Local
	Target Local
}

// AssignVarOnceStmt represents an assignment of one local variable to another.
// If the target is defined, execution aborts with a conflict error.
//
// TODO(tsandall): is there a better name for this?
type AssignVarOnceStmt struct {
	Target Local
	Source Local
}

// MakeStringStmt constructs a local variable that refers to a string constant.
type MakeStringStmt struct {
	Index  int
	Target Local
}

// MakeNullStmt constructs a local variable that refers to a null value.
type MakeNullStmt struct {
	Target Local
}

// MakeBooleanStmt constructs a local variable that refers to a boolean value.
type MakeBooleanStmt struct {
	Value  bool
	Target Local
}

// MakeNumberFloatStmt constructs a local variable that refers to a
// floating-point number value.
type MakeNumberFloatStmt struct {
	Value  float64
	Target Local
}

// MakeNumberIntStmt constructs a local variable that refers to an integer value.
type MakeNumberIntStmt struct {
	Value  int64
	Target Local
}

// MakeNumberRefStmt constructs a local variable that refers to a number stored as a string.
type MakeNumberRefStmt struct {
	Index  int
	Target Local
}

// MakeArrayStmt constructs a local variable that refers to an array value.
type MakeArrayStmt struct {
	Capacity int32
	Target   Local
}

// MakeObjectStmt constructs a local variable that refers to an object value.
type MakeObjectStmt struct {
	Target Local
}

// MakeSetStmt constructs a local variable that refers to a set value.
type MakeSetStmt struct {
	Target Local
}

// EqualStmt represents an value-equality check of two local variables.
type EqualStmt struct {
	A Local
	B Local
}

// LessThanStmt represents a < check of two local variables.
type LessThanStmt struct {
	A Local
	B Local
}

// LessThanEqualStmt represents a <= check of two local variables.
type LessThanEqualStmt struct {
	A Local
	B Local
}

// GreaterThanStmt represents a > check of two local variables.
type GreaterThanStmt struct {
	A Local
	B Local
}

// GreaterThanEqualStmt represents a >= check of two local variables.
type GreaterThanEqualStmt struct {
	A Local
	B Local
}

// NotEqualStmt represents a != check of two local variables.
type NotEqualStmt struct {
	A Local
	B Local
}

// IsArrayStmt represents a dynamic type check on a local variable.
type IsArrayStmt struct {
	Source Local
}

// IsObjectStmt represents a dynamic type check on a local variable.
type IsObjectStmt struct {
	Source Local
}

// IsDefinedStmt represents a check of whether a local variable is defined.
type IsDefinedStmt struct {
	Source Local
}

// IsUndefinedStmt represents a check of whether local variable is undefined.
type IsUndefinedStmt struct {
	Source Local
}

// ArrayAppendStmt represents a dynamic append operation of a value
// onto an array.
type ArrayAppendStmt struct {
	Value Local
	Array Local
}

// ObjectInsertStmt represents a dynamic insert operation of a
// key/value pair into an object.
type ObjectInsertStmt struct {
	Key    Local
	Value  Local
	Object Local
}

// ObjectInsertOnceStmt represents a dynamic insert operation of a key/value
// pair into an object. If the key already exists and the value differs,
// execution aborts with a conflict error.
type ObjectInsertOnceStmt struct {
	Key    Local
	Value  Local
	Object Local
}

// ObjectMergeStmt performs a recursive merge of two object values. If either of
// the locals refer to non-object values this operation will abort with a
// conflict error. Overlapping object keys are merged recursively.
type ObjectMergeStmt struct {
	A      Local
	B      Local
	Target Local
}

// SetAddStmt represents a dynamic add operation of an element into a set.
type SetAddStmt struct {
	Value Local
	Set   Local
}

// WithStmt replaces the Local or a portion of the document referred to by the
// Local with the Value and executes the contained block. If the Path is
// non-empty, the Value is upserted into the Local. If the intermediate nodes in
// the Local referred to by the Path do not exist, they will be created. When
// the WithStmt finishes the Local is reset to it's original value.
type WithStmt struct {
	Local Local
	Path  []int
	Value Local
	Block *Block
}

// ResultSetAdd adds a value into the result set returned by the query plan.
type ResultSetAdd struct {
	Value Local
}
