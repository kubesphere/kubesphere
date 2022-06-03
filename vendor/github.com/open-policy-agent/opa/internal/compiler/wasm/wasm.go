// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package wasm contains an IR->WASM compiler backend.
package wasm

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/internal/compiler/wasm/opa"
	"github.com/open-policy-agent/opa/internal/ir"
	"github.com/open-policy-agent/opa/internal/wasm/encoding"
	"github.com/open-policy-agent/opa/internal/wasm/instruction"
	"github.com/open-policy-agent/opa/internal/wasm/module"
	"github.com/open-policy-agent/opa/internal/wasm/types"
)

const (
	opaTypeNull int32 = iota + 1
	opaTypeBoolean
	opaTypeNumber
	opaTypeString
	opaTypeArray
	opaTypeObject
)

const (
	opaFuncPrefix        = "opa_"
	opaAbort             = "opa_abort"
	opaJSONParse         = "opa_json_parse"
	opaNull              = "opa_null"
	opaBoolean           = "opa_boolean"
	opaNumberInt         = "opa_number_int"
	opaNumberFloat       = "opa_number_float"
	opaNumberRef         = "opa_number_ref"
	opaNumberSize        = "opa_number_size"
	opaArrayWithCap      = "opa_array_with_cap"
	opaArrayAppend       = "opa_array_append"
	opaObject            = "opa_object"
	opaObjectInsert      = "opa_object_insert"
	opaSet               = "opa_set"
	opaSetAdd            = "opa_set_add"
	opaStringTerminated  = "opa_string_terminated"
	opaValueBooleanSet   = "opa_value_boolean_set"
	opaValueNumberSetInt = "opa_value_number_set_int"
	opaValueCompare      = "opa_value_compare"
	opaValueGet          = "opa_value_get"
	opaValueIter         = "opa_value_iter"
	opaValueLength       = "opa_value_length"
	opaValueMerge        = "opa_value_merge"
	opaValueShallowCopy  = "opa_value_shallow_copy"
	opaValueType         = "opa_value_type"
)

var builtinsFunctions = map[string]string{
	ast.Plus.Name:            "opa_arith_plus",
	ast.Minus.Name:           "opa_arith_minus",
	ast.Multiply.Name:        "opa_arith_multiply",
	ast.Divide.Name:          "opa_arith_divide",
	ast.Abs.Name:             "opa_arith_abs",
	ast.Round.Name:           "opa_arith_round",
	ast.Rem.Name:             "opa_arith_rem",
	ast.ArrayConcat.Name:     "opa_array_concat",
	ast.ArraySlice.Name:      "opa_array_slice",
	ast.SetDiff.Name:         "opa_set_diff",
	ast.And.Name:             "opa_set_intersection",
	ast.Or.Name:              "opa_set_union",
	ast.Intersection.Name:    "opa_sets_intersection",
	ast.Union.Name:           "opa_sets_union",
	ast.IsNumber.Name:        "opa_types_is_number",
	ast.IsString.Name:        "opa_types_is_string",
	ast.IsBoolean.Name:       "opa_types_is_boolean",
	ast.IsArray.Name:         "opa_types_is_array",
	ast.IsSet.Name:           "opa_types_is_set",
	ast.IsObject.Name:        "opa_types_is_object",
	ast.IsNull.Name:          "opa_types_is_null",
	ast.TypeNameBuiltin.Name: "opa_types_name",
	ast.BitsOr.Name:          "opa_bits_or",
	ast.BitsAnd.Name:         "opa_bits_and",
	ast.BitsNegate.Name:      "opa_bits_negate",
	ast.BitsXOr.Name:         "opa_bits_xor",
	ast.BitsShiftLeft.Name:   "opa_bits_shiftleft",
	ast.BitsShiftRight.Name:  "opa_bits_shiftright",
	ast.Count.Name:           "opa_agg_count",
	ast.Sum.Name:             "opa_agg_sum",
	ast.Product.Name:         "opa_agg_product",
	ast.Max.Name:             "opa_agg_max",
	ast.Min.Name:             "opa_agg_min",
	ast.Sort.Name:            "opa_agg_sort",
	ast.All.Name:             "opa_agg_all",
	ast.Any.Name:             "opa_agg_any",
	ast.Concat.Name:          "opa_strings_concat",
	ast.FormatInt.Name:       "opa_strings_format_int",
	ast.IndexOf.Name:         "opa_strings_indexof",
	ast.Substring.Name:       "opa_strings_substring",
	ast.Lower.Name:           "opa_strings_lower",
	ast.Upper.Name:           "opa_strings_upper",
	ast.Contains.Name:        "opa_strings_contains",
	ast.StartsWith.Name:      "opa_strings_startswith",
	ast.EndsWith.Name:        "opa_strings_endswith",
	ast.Split.Name:           "opa_strings_split",
	ast.Replace.Name:         "opa_strings_replace",
	ast.ReplaceN.Name:        "opa_strings_replace_n",
	ast.Trim.Name:            "opa_strings_trim",
	ast.TrimLeft.Name:        "opa_strings_trim_left",
	ast.TrimPrefix.Name:      "opa_strings_trim_prefix",
	ast.TrimRight.Name:       "opa_strings_trim_right",
	ast.TrimSuffix.Name:      "opa_strings_trim_suffix",
	ast.TrimSpace.Name:       "opa_strings_trim_space",
	ast.NumbersRange.Name:    "opa_numbers_range",
}

var builtinDispatchers = [...]string{
	"opa_builtin0",
	"opa_builtin1",
	"opa_builtin2",
	"opa_builtin3",
	"opa_builtin4",
}

// Compiler implements an IR->WASM compiler backend.
type Compiler struct {
	stages []func() error // compiler stages to execute
	errors []error        // compilation errors encountered

	policy *ir.Policy        // input policy to compile
	module *module.Module    // output WASM module
	code   *module.CodeEntry // output WASM code

	builtinStringAddrs    map[int]uint32    // addresses of built-in string constants
	externalFuncNameAddrs map[string]int32  // addresses of required built-in function names for listing
	externalFuncs         map[string]int32  // required built-in function ids
	stringOffset          int32             // null-terminated string data base offset
	stringAddrs           []uint32          // null-terminated string constant addresses
	funcs                 map[string]uint32 // maps imported and exported function names to function indices

	nextLocal uint32
	locals    map[ir.Local]uint32
	lctx      uint32 // local pointing to eval context
	lrs       uint32 // local pointing to result set
}

const (
	errVarAssignConflict int = iota
	errObjectInsertConflict
	errObjectMergeConflict
	errWithConflict
)

var errorMessages = [...]struct {
	id      int
	message string
}{
	{errVarAssignConflict, "var assignment conflict"},
	{errObjectInsertConflict, "object insert conflict"},
	{errObjectMergeConflict, "object merge conflict"},
	{errWithConflict, "with target conflict"},
}

// New returns a new compiler object.
func New() *Compiler {
	c := &Compiler{}
	c.stages = []func() error{
		c.initModule,
		c.compileStrings,
		c.compileExternalFuncDecls,
		c.compileFuncs,
		c.compilePlan,
	}
	return c
}

// WithPolicy sets the policy to compile.
func (c *Compiler) WithPolicy(p *ir.Policy) *Compiler {
	c.policy = p
	return c
}

// Compile returns a compiled WASM module.
func (c *Compiler) Compile() (*module.Module, error) {

	for _, stage := range c.stages {
		if err := stage(); err != nil {
			return nil, err
		} else if len(c.errors) > 0 {
			return nil, c.errors[0] // TODO(tsandall) return all errors.
		}
	}

	return c.module, nil
}

// initModule instantiates the module from the pre-compiled OPA binary. The
// module is then updated to include declarations for all of the functions that
// are about to be compiled.
func (c *Compiler) initModule() error {

	bs, err := opa.Bytes()
	if err != nil {
		return err
	}

	c.module, err = encoding.ReadModule(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	c.funcs = make(map[string]uint32)
	var funcidx uint32

	for _, imp := range c.module.Import.Imports {
		if imp.Descriptor.Kind() == module.FunctionImportType {
			c.funcs[imp.Name] = funcidx
			funcidx++
		}
	}

	for i := range c.module.Export.Exports {
		exp := &c.module.Export.Exports[i]
		if exp.Descriptor.Type == module.FunctionExportType {
			c.funcs[exp.Name] = exp.Descriptor.Index
		}
	}

	for _, fn := range c.policy.Funcs.Funcs {

		params := make([]types.ValueType, len(fn.Params))
		for i := 0; i < len(params); i++ {
			params[i] = types.I32
		}

		tpe := module.FunctionType{
			Params:  params,
			Results: []types.ValueType{types.I32},
		}

		c.emitFunctionDecl(fn.Name, tpe, false)
	}

	c.emitFunctionDecl("eval", module.FunctionType{
		Params:  []types.ValueType{types.I32},
		Results: []types.ValueType{types.I32},
	}, true)

	c.emitFunctionDecl("builtins", module.FunctionType{
		Params:  nil,
		Results: []types.ValueType{types.I32},
	}, true)

	return nil
}

// compileStrings compiles string constants into the data section of the module.
// The strings are indexed for lookups in later stages.
func (c *Compiler) compileStrings() error {

	var err error
	c.stringOffset, err = getLowestFreeDataSegmentOffset(c.module)
	if err != nil {
		return err
	}

	c.stringAddrs = make([]uint32, len(c.policy.Static.Strings))
	var buf bytes.Buffer

	for i, s := range c.policy.Static.Strings {
		addr := uint32(buf.Len()) + uint32(c.stringOffset)
		buf.WriteString(s.Value)
		buf.WriteByte(0)
		c.stringAddrs[i] = addr
	}

	c.externalFuncNameAddrs = make(map[string]int32)

	for _, decl := range c.policy.Static.BuiltinFuncs {
		if _, ok := builtinsFunctions[decl.Name]; !ok {
			addr := int32(buf.Len()) + int32(c.stringOffset)
			buf.WriteString(decl.Name)
			buf.WriteByte(0)
			c.externalFuncNameAddrs[decl.Name] = addr
		}
	}

	c.builtinStringAddrs = make(map[int]uint32, len(errorMessages))

	for i := range errorMessages {
		addr := uint32(buf.Len()) + uint32(c.stringOffset)
		buf.WriteString(errorMessages[i].message)
		buf.WriteByte(0)
		c.builtinStringAddrs[errorMessages[i].id] = addr
	}

	c.module.Data.Segments = append(c.module.Data.Segments, module.DataSegment{
		Index: 0,
		Offset: module.Expr{
			Instrs: []instruction.Instruction{
				instruction.I32Const{
					Value: c.stringOffset,
				},
			},
		},
		Init: buf.Bytes(),
	})

	return nil
}

// compileExternalFuncDecls generates a function that lists the built-ins required by
// the policy. The host environment should invoke this function obtain the list
// of built-in function identifiers (represented as integers) that will be used
// when calling out.
func (c *Compiler) compileExternalFuncDecls() error {

	c.code = &module.CodeEntry{}
	c.nextLocal = 0
	c.locals = map[ir.Local]uint32{}

	lobj := c.genLocal()

	c.appendInstr(instruction.Call{Index: c.function(opaObject)})
	c.appendInstr(instruction.SetLocal{Index: lobj})
	c.externalFuncs = make(map[string]int32)

	for index, decl := range c.policy.Static.BuiltinFuncs {
		if _, ok := builtinsFunctions[decl.Name]; !ok {
			c.appendInstr(instruction.GetLocal{Index: lobj})
			c.appendInstr(instruction.I32Const{Value: c.externalFuncNameAddrs[decl.Name]})
			c.appendInstr(instruction.Call{Index: c.function(opaStringTerminated)})
			c.appendInstr(instruction.I64Const{Value: int64(index)})
			c.appendInstr(instruction.Call{Index: c.function(opaNumberInt)})
			c.appendInstr(instruction.Call{Index: c.function(opaObjectInsert)})
			c.externalFuncs[decl.Name] = int32(index)
		}
	}

	c.appendInstr(instruction.GetLocal{Index: lobj})

	c.code.Func.Locals = []module.LocalDeclaration{
		{
			Count: c.nextLocal,
			Type:  types.I32,
		},
	}

	return c.emitFunction("builtins", c.code)
}

// compileFuncs compiles the policy functions and emits them into the module.
func (c *Compiler) compileFuncs() error {

	for _, fn := range c.policy.Funcs.Funcs {
		if err := c.compileFunc(fn); err != nil {
			return errors.Wrapf(err, "func %v", fn.Name)
		}
	}

	return nil
}

// compilePlan compiles the policy plan and emits the resulting function into
// the module.
func (c *Compiler) compilePlan() error {

	c.code = &module.CodeEntry{}
	c.nextLocal = 0
	c.locals = map[ir.Local]uint32{}
	c.lctx = c.genLocal()
	c.lrs = c.genLocal()

	// Initialize the input and data locals.
	c.appendInstr(instruction.GetLocal{Index: c.lctx})
	c.appendInstr(instruction.I32Load{Offset: 0, Align: 2})
	c.appendInstr(instruction.SetLocal{Index: c.local(ir.Input)})

	c.appendInstr(instruction.GetLocal{Index: c.lctx})
	c.appendInstr(instruction.I32Load{Offset: 4, Align: 2})
	c.appendInstr(instruction.SetLocal{Index: c.local(ir.Data)})

	// Initialize the result set.
	c.appendInstr(instruction.Call{Index: c.function(opaSet)})
	c.appendInstr(instruction.SetLocal{Index: c.lrs})
	c.appendInstr(instruction.GetLocal{Index: c.lctx})
	c.appendInstr(instruction.GetLocal{Index: c.lrs})
	c.appendInstr(instruction.I32Store{Offset: 8, Align: 2})

	for i := range c.policy.Plan.Blocks {

		instrs, err := c.compileBlock(c.policy.Plan.Blocks[i])
		if err != nil {
			return errors.Wrapf(err, "block %d", i)
		}

		c.appendInstr(instruction.Block{Instrs: instrs})
	}

	c.appendInstr(instruction.I32Const{Value: int32(0)})

	c.code.Func.Locals = []module.LocalDeclaration{
		{
			Count: c.nextLocal,
			Type:  types.I32,
		},
	}

	return c.emitFunction("eval", c.code)
}

func (c *Compiler) compileFunc(fn *ir.Func) error {

	if len(fn.Params) == 0 {
		return fmt.Errorf("illegal function: zero args")
	}

	c.nextLocal = 0
	c.locals = map[ir.Local]uint32{}

	for _, a := range fn.Params {
		_ = c.local(a)
	}

	_ = c.local(fn.Return)

	c.code = &module.CodeEntry{}

	for i := range fn.Blocks {
		instrs, err := c.compileBlock(fn.Blocks[i])
		if err != nil {
			return errors.Wrapf(err, "block %d", i)
		}
		if i < len(fn.Blocks)-1 {
			c.appendInstr(instruction.Block{Instrs: instrs})
		} else {
			c.appendInstrs(instrs)
		}
	}

	c.code.Func.Locals = []module.LocalDeclaration{
		{
			Count: c.nextLocal,
			Type:  types.I32,
		},
	}

	var params []types.ValueType

	for i := 0; i < len(fn.Params); i++ {
		params = append(params, types.I32)
	}

	return c.emitFunction(fn.Name, c.code)
}

func (c *Compiler) compileBlock(block *ir.Block) ([]instruction.Instruction, error) {

	var instrs []instruction.Instruction

	for _, stmt := range block.Stmts {
		switch stmt := stmt.(type) {
		case *ir.ResultSetAdd:
			instrs = append(instrs, instruction.GetLocal{Index: c.lrs})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Value)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaSetAdd)})
		case *ir.ReturnLocalStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.Return{})
		case *ir.BlockStmt:
			for i := range stmt.Blocks {
				block, err := c.compileBlock(stmt.Blocks[i])
				if err != nil {
					return nil, err
				}
				instrs = append(instrs, instruction.Block{Instrs: block})
			}
		case *ir.BreakStmt:
			instrs = append(instrs, instruction.Br{Index: stmt.Index})
		case *ir.CallStmt:
			if err := c.compileCallStmt(stmt, &instrs); err != nil {
				return nil, err
			}
		case *ir.WithStmt:
			if err := c.compileWithStmt(stmt, &instrs); err != nil {
				return instrs, err
			}
		case *ir.AssignVarStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.AssignVarOnceStmt:
			instrs = append(instrs, instruction.Block{
				Instrs: []instruction.Instruction{
					instruction.Block{
						Instrs: []instruction.Instruction{
							instruction.GetLocal{Index: c.local(stmt.Target)},
							instruction.I32Eqz{},
							instruction.BrIf{Index: 0},
							instruction.GetLocal{Index: c.local(stmt.Target)},
							instruction.GetLocal{Index: c.local(stmt.Source)},
							instruction.Call{Index: c.function(opaValueCompare)},
							instruction.I32Eqz{},
							instruction.BrIf{Index: 1},
							instruction.I32Const{Value: c.builtinStringAddr(errVarAssignConflict)},
							instruction.Call{Index: c.function(opaAbort)},
							instruction.Unreachable{},
						},
					},
					instruction.GetLocal{Index: c.local(stmt.Source)},
					instruction.SetLocal{Index: c.local(stmt.Target)},
				},
			})
		case *ir.AssignBooleanStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Target)})
			if stmt.Value {
				instrs = append(instrs, instruction.I32Const{Value: 1})
			} else {
				instrs = append(instrs, instruction.I32Const{Value: 0})
			}
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueBooleanSet)})
		case *ir.AssignIntStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Target)})
			instrs = append(instrs, instruction.I64Const{Value: stmt.Value})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueNumberSetInt)})
		case *ir.ScanStmt:
			if err := c.compileScan(stmt, &instrs); err != nil {
				return nil, err
			}
		case *ir.NotStmt:
			if err := c.compileNot(stmt, &instrs); err != nil {
				return nil, err
			}
		case *ir.DotStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Key)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueGet)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Target)})
			instrs = append(instrs, instruction.I32Eqz{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.LenStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueLength)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaNumberSize)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.EqualStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.LessThanStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.I32Const{Value: 0})
			instrs = append(instrs, instruction.I32GeS{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.LessThanEqualStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.I32Const{Value: 0})
			instrs = append(instrs, instruction.I32GtS{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.GreaterThanStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.I32Const{Value: 0})
			instrs = append(instrs, instruction.I32LeS{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.GreaterThanEqualStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.I32Const{Value: 0})
			instrs = append(instrs, instruction.I32LtS{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.NotEqualStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueCompare)})
			instrs = append(instrs, instruction.I32Eqz{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.MakeNullStmt:
			instrs = append(instrs, instruction.Call{Index: c.function(opaNull)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeBooleanStmt:
			instr := instruction.I32Const{}
			if stmt.Value {
				instr.Value = 1
			} else {
				instr.Value = 0
			}
			instrs = append(instrs, instr)
			instrs = append(instrs, instruction.Call{Index: c.function(opaBoolean)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeNumberFloatStmt:
			instrs = append(instrs, instruction.F64Const{Value: stmt.Value})
			instrs = append(instrs, instruction.Call{Index: c.function(opaNumberFloat)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeNumberIntStmt:
			instrs = append(instrs, instruction.I64Const{Value: stmt.Value})
			instrs = append(instrs, instruction.Call{Index: c.function(opaNumberInt)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeNumberRefStmt:
			instrs = append(instrs, instruction.I32Const{Value: c.stringAddr(stmt.Index)})
			instrs = append(instrs, instruction.I32Const{Value: int32(len(c.policy.Static.Strings[stmt.Index].Value))})
			instrs = append(instrs, instruction.Call{Index: c.function(opaNumberRef)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeStringStmt:
			instrs = append(instrs, instruction.I32Const{Value: c.stringAddr(stmt.Index)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaStringTerminated)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeArrayStmt:
			instrs = append(instrs, instruction.I32Const{Value: stmt.Capacity})
			instrs = append(instrs, instruction.Call{Index: c.function(opaArrayWithCap)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeObjectStmt:
			instrs = append(instrs, instruction.Call{Index: c.function(opaObject)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.MakeSetStmt:
			instrs = append(instrs, instruction.Call{Index: c.function(opaSet)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
		case *ir.IsArrayStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueType)})
			instrs = append(instrs, instruction.I32Const{Value: opaTypeArray})
			instrs = append(instrs, instruction.I32Ne{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.IsObjectStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueType)})
			instrs = append(instrs, instruction.I32Const{Value: opaTypeObject})
			instrs = append(instrs, instruction.I32Ne{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.IsUndefinedStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.I32Const{Value: 0})
			instrs = append(instrs, instruction.I32Ne{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.IsDefinedStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Source)})
			instrs = append(instrs, instruction.I32Eqz{})
			instrs = append(instrs, instruction.BrIf{Index: 0})
		case *ir.ArrayAppendStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Array)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Value)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaArrayAppend)})
		case *ir.ObjectInsertStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Object)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Key)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Value)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaObjectInsert)})
		case *ir.ObjectInsertOnceStmt:
			tmp := c.genLocal()
			instrs = append(instrs, instruction.Block{
				Instrs: []instruction.Instruction{
					instruction.Block{
						Instrs: []instruction.Instruction{
							instruction.GetLocal{Index: c.local(stmt.Object)},
							instruction.GetLocal{Index: c.local(stmt.Key)},
							instruction.Call{Index: c.function(opaValueGet)},
							instruction.SetLocal{Index: tmp},
							instruction.GetLocal{Index: tmp},
							instruction.I32Eqz{},
							instruction.BrIf{Index: 0},
							instruction.GetLocal{Index: tmp},
							instruction.GetLocal{Index: c.local(stmt.Value)},
							instruction.Call{Index: c.function(opaValueCompare)},
							instruction.I32Eqz{},
							instruction.BrIf{Index: 1},
							instruction.I32Const{Value: c.builtinStringAddr(errObjectInsertConflict)},
							instruction.Call{Index: c.function(opaAbort)},
							instruction.Unreachable{},
						},
					},
					instruction.GetLocal{Index: c.local(stmt.Object)},
					instruction.GetLocal{Index: c.local(stmt.Key)},
					instruction.GetLocal{Index: c.local(stmt.Value)},
					instruction.Call{Index: c.function(opaObjectInsert)},
				},
			})
		case *ir.ObjectMergeStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.A)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.B)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaValueMerge)})
			instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Target)})
			instrs = append(instrs, instruction.Block{
				Instrs: []instruction.Instruction{
					instruction.GetLocal{Index: c.local(stmt.Target)},
					instruction.BrIf{Index: 0},
					instruction.I32Const{Value: c.builtinStringAddr(errObjectMergeConflict)},
					instruction.Call{Index: c.function(opaAbort)},
					instruction.Unreachable{},
				},
			})
		case *ir.SetAddStmt:
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Set)})
			instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Value)})
			instrs = append(instrs, instruction.Call{Index: c.function(opaSetAdd)})
		default:
			var buf bytes.Buffer
			ir.Pretty(&buf, stmt)
			return instrs, fmt.Errorf("illegal statement: %v", buf.String())
		}
	}

	return instrs, nil
}

func (c *Compiler) compileScan(scan *ir.ScanStmt, result *[]instruction.Instruction) error {
	var instrs = *result
	instrs = append(instrs, instruction.I32Const{Value: 0})
	instrs = append(instrs, instruction.SetLocal{Index: c.local(scan.Key)})
	body, err := c.compileScanBlock(scan)
	if err != nil {
		return err
	}
	instrs = append(instrs, instruction.Block{
		Instrs: []instruction.Instruction{
			instruction.Loop{Instrs: body},
		},
	})
	*result = instrs
	return nil
}

func (c *Compiler) compileScanBlock(scan *ir.ScanStmt) ([]instruction.Instruction, error) {
	var instrs []instruction.Instruction

	// Execute iterator.
	instrs = append(instrs, instruction.GetLocal{Index: c.local(scan.Source)})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(scan.Key)})
	instrs = append(instrs, instruction.Call{Index: c.function(opaValueIter)})

	// Check for emptiness.
	instrs = append(instrs, instruction.SetLocal{Index: c.local(scan.Key)})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(scan.Key)})
	instrs = append(instrs, instruction.I32Eqz{})
	instrs = append(instrs, instruction.BrIf{Index: 1})

	// Load value.
	instrs = append(instrs, instruction.GetLocal{Index: c.local(scan.Source)})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(scan.Key)})
	instrs = append(instrs, instruction.Call{Index: c.function(opaValueGet)})
	instrs = append(instrs, instruction.SetLocal{Index: c.local(scan.Value)})

	// Loop body.
	nested, err := c.compileBlock(scan.Block)
	if err != nil {
		return nil, err
	}

	// Continue.
	instrs = append(instrs, nested...)
	instrs = append(instrs, instruction.Br{Index: 0})

	return instrs, nil
}

func (c *Compiler) compileNot(not *ir.NotStmt, result *[]instruction.Instruction) error {
	var instrs = *result

	// generate and initialize condition variable
	cond := c.genLocal()
	instrs = append(instrs, instruction.I32Const{Value: 1})
	instrs = append(instrs, instruction.SetLocal{Index: cond})

	nested, err := c.compileBlock(not.Block)
	if err != nil {
		return err
	}

	// unset condition variable if end of block is reached
	nested = append(nested, instruction.I32Const{Value: 0})
	nested = append(nested, instruction.SetLocal{Index: cond})
	instrs = append(instrs, instruction.Block{Instrs: nested})

	// break out of block if condition variable was unset
	instrs = append(instrs, instruction.GetLocal{Index: cond})
	instrs = append(instrs, instruction.I32Eqz{})
	instrs = append(instrs, instruction.BrIf{Index: 0})

	*result = instrs
	return nil
}

func (c *Compiler) compileWithStmt(with *ir.WithStmt, result *[]instruction.Instruction) error {

	var instrs = *result
	save := c.genLocal()
	instrs = append(instrs, instruction.GetLocal{Index: c.local(with.Local)})
	instrs = append(instrs, instruction.SetLocal{Index: save})

	if len(with.Path) == 0 {
		instrs = append(instrs, instruction.GetLocal{Index: c.local(with.Value)})
		instrs = append(instrs, instruction.SetLocal{Index: c.local(with.Local)})
	} else {
		instrs = c.compileUpsert(with.Local, with.Path, with.Value, instrs)
	}

	undefined := c.genLocal()
	instrs = append(instrs, instruction.I32Const{Value: 1})
	instrs = append(instrs, instruction.SetLocal{Index: undefined})

	nested, err := c.compileBlock(with.Block)
	if err != nil {
		return err
	}

	nested = append(nested, instruction.I32Const{Value: 0})
	nested = append(nested, instruction.SetLocal{Index: undefined})
	instrs = append(instrs, instruction.Block{Instrs: nested})
	instrs = append(instrs, instruction.GetLocal{Index: save})
	instrs = append(instrs, instruction.SetLocal{Index: c.local(with.Local)})
	instrs = append(instrs, instruction.GetLocal{Index: undefined})
	instrs = append(instrs, instruction.BrIf{Index: 0})

	*result = instrs

	return nil
}

func (c *Compiler) compileUpsert(local ir.Local, path []int, value ir.Local, instrs []instruction.Instruction) []instruction.Instruction {

	lcopy := c.genLocal() // holds copy of local
	instrs = append(instrs, instruction.GetLocal{Index: c.local(local)})
	instrs = append(instrs, instruction.SetLocal{Index: lcopy})

	// Shallow copy the local if defined otherwise initialize to an empty object.
	instrs = append(instrs, instruction.Block{
		Instrs: []instruction.Instruction{
			instruction.Block{Instrs: []instruction.Instruction{
				instruction.GetLocal{Index: lcopy},
				instruction.I32Eqz{},
				instruction.BrIf{Index: 0},
				instruction.GetLocal{Index: lcopy},
				instruction.Call{Index: c.function(opaValueShallowCopy)},
				instruction.SetLocal{Index: lcopy},
				instruction.GetLocal{Index: lcopy},
				instruction.SetLocal{Index: c.local(local)},
				instruction.Br{Index: 1},
			}},
			instruction.Call{Index: c.function(opaObject)},
			instruction.SetLocal{Index: lcopy},
			instruction.GetLocal{Index: lcopy},
			instruction.SetLocal{Index: c.local(local)},
		},
	})

	// Initialize the locals that specify the path of the upsert operation.
	lpath := make(map[int]uint32, len(path))

	for i := 0; i < len(path); i++ {
		lpath[i] = c.genLocal()
		instrs = append(instrs, instruction.I32Const{Value: c.stringAddr(path[i])})
		instrs = append(instrs, instruction.Call{Index: c.function(opaStringTerminated)})
		instrs = append(instrs, instruction.SetLocal{Index: lpath[i]})
	}

	// Generate a block that traverses the path of the upsert operation,
	// shallowing copying values at each step as needed. Stop before the final
	// segment that will only be inserted.
	var inner []instruction.Instruction
	ltemp := c.genLocal()

	for i := 0; i < len(path)-1; i++ {

		// Lookup the next part of the path.
		inner = append(inner, instruction.GetLocal{Index: lcopy})
		inner = append(inner, instruction.GetLocal{Index: lpath[i]})
		inner = append(inner, instruction.Call{Index: c.function(opaValueGet)})
		inner = append(inner, instruction.SetLocal{Index: ltemp})

		// If the next node is missing, break.
		inner = append(inner, instruction.GetLocal{Index: ltemp})
		inner = append(inner, instruction.I32Eqz{})
		inner = append(inner, instruction.BrIf{Index: uint32(i)})

		// If the next node is not an object, generate a conflict error.
		inner = append(inner, instruction.Block{
			Instrs: []instruction.Instruction{
				instruction.GetLocal{Index: ltemp},
				instruction.Call{Index: c.function(opaValueType)},
				instruction.I32Const{Value: opaTypeObject},
				instruction.I32Eq{},
				instruction.BrIf{Index: 0},
				instruction.I32Const{Value: c.builtinStringAddr(errWithConflict)},
				instruction.Call{Index: c.function(opaAbort)},
			},
		})

		// Otherwise, shallow copy the next node node and insert into the copy
		// before continuing.
		inner = append(inner, instruction.GetLocal{Index: ltemp})
		inner = append(inner, instruction.Call{Index: c.function(opaValueShallowCopy)})
		inner = append(inner, instruction.SetLocal{Index: ltemp})
		inner = append(inner, instruction.GetLocal{Index: lcopy})
		inner = append(inner, instruction.GetLocal{Index: lpath[i]})
		inner = append(inner, instruction.GetLocal{Index: ltemp})
		inner = append(inner, instruction.Call{Index: c.function(opaObjectInsert)})
		inner = append(inner, instruction.GetLocal{Index: ltemp})
		inner = append(inner, instruction.SetLocal{Index: lcopy})
	}

	inner = append(inner, instruction.Br{Index: uint32(len(path) - 1)})

	// Generate blocks that handle missing nodes during traversal.
	var block []instruction.Instruction
	lval := c.genLocal()

	for i := 0; i < len(path)-1; i++ {
		block = append(block, instruction.Block{Instrs: inner})
		block = append(block, instruction.Call{Index: c.function(opaObject)})
		block = append(block, instruction.SetLocal{Index: lval})
		block = append(block, instruction.GetLocal{Index: lcopy})
		block = append(block, instruction.GetLocal{Index: lpath[i]})
		block = append(block, instruction.GetLocal{Index: lval})
		block = append(block, instruction.Call{Index: c.function(opaObjectInsert)})
		block = append(block, instruction.GetLocal{Index: lval})
		block = append(block, instruction.SetLocal{Index: lcopy})
		inner = block
		block = nil
	}

	// Finish by inserting the statement's value into the shallow copied node.
	instrs = append(instrs, instruction.Block{Instrs: inner})
	instrs = append(instrs, instruction.GetLocal{Index: lcopy})
	instrs = append(instrs, instruction.GetLocal{Index: lpath[len(path)-1]})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(value)})
	instrs = append(instrs, instruction.Call{Index: c.function(opaObjectInsert)})

	return instrs
}

func (c *Compiler) compileCallStmt(stmt *ir.CallStmt, result *[]instruction.Instruction) error {

	fn := stmt.Func

	if name, ok := builtinsFunctions[stmt.Func]; ok {
		fn = name
	}

	if index, ok := c.funcs[fn]; ok {
		return c.compileInternalCall(stmt, index, result)
	}

	if id, ok := c.externalFuncs[fn]; ok {
		return c.compileExternalCall(stmt, id, result)
	}

	c.errors = append(c.errors, fmt.Errorf("undefined function: %q", fn))

	return nil
}

func (c *Compiler) compileInternalCall(stmt *ir.CallStmt, index uint32, result *[]instruction.Instruction) error {
	instrs := *result

	for _, arg := range stmt.Args {
		instrs = append(instrs, instruction.GetLocal{Index: c.local(arg)})
	}

	instrs = append(instrs, instruction.Call{Index: index})
	instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Result)})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Result)})
	instrs = append(instrs, instruction.I32Eqz{})
	instrs = append(instrs, instruction.BrIf{Index: 0})
	*result = instrs
	return nil
}

func (c *Compiler) compileExternalCall(stmt *ir.CallStmt, id int32, result *[]instruction.Instruction) error {

	if len(stmt.Args) >= len(builtinDispatchers) {
		c.errors = append(c.errors, fmt.Errorf("too many built-in call arguments: %q", stmt.Func))
		return nil
	}

	instrs := *result
	instrs = append(instrs, instruction.I32Const{Value: id})
	instrs = append(instrs, instruction.I32Const{Value: 0}) // unused context parameter

	for _, arg := range stmt.Args {
		instrs = append(instrs, instruction.GetLocal{Index: c.local(arg)})
	}

	instrs = append(instrs, instruction.Call{Index: c.funcs[builtinDispatchers[len(stmt.Args)]]})
	instrs = append(instrs, instruction.SetLocal{Index: c.local(stmt.Result)})
	instrs = append(instrs, instruction.GetLocal{Index: c.local(stmt.Result)})
	instrs = append(instrs, instruction.I32Eqz{})
	instrs = append(instrs, instruction.BrIf{Index: 0})
	*result = instrs
	return nil
}

func (c *Compiler) emitFunctionDecl(name string, tpe module.FunctionType, export bool) {

	typeIndex := c.emitFunctionType(tpe)
	c.module.Function.TypeIndices = append(c.module.Function.TypeIndices, typeIndex)
	c.module.Code.Segments = append(c.module.Code.Segments, module.RawCodeSegment{})
	c.funcs[name] = uint32((len(c.module.Function.TypeIndices) - 1) + c.functionImportCount())

	if export {
		c.module.Export.Exports = append(c.module.Export.Exports, module.Export{
			Name: name,
			Descriptor: module.ExportDescriptor{
				Type:  module.FunctionExportType,
				Index: c.funcs[name],
			},
		})
	}

}

func (c *Compiler) emitFunctionType(tpe module.FunctionType) uint32 {
	for i, other := range c.module.Type.Functions {
		if tpe.Equal(other) {
			return uint32(i)
		}
	}
	c.module.Type.Functions = append(c.module.Type.Functions, tpe)
	return uint32(len(c.module.Type.Functions) - 1)
}

func (c *Compiler) emitFunction(name string, entry *module.CodeEntry) error {
	var buf bytes.Buffer
	if err := encoding.WriteCodeEntry(&buf, entry); err != nil {
		return err
	}
	index := c.function(name) - uint32(c.functionImportCount())
	c.module.Code.Segments[index].Code = buf.Bytes()
	return nil
}

func (c *Compiler) functionImportCount() int {
	var count int

	for _, imp := range c.module.Import.Imports {
		if imp.Descriptor.Kind() == module.FunctionImportType {
			count++
		}
	}

	return count
}

func (c *Compiler) stringAddr(index int) int32 {
	return int32(c.stringAddrs[index])
}

func (c *Compiler) builtinStringAddr(code int) int32 {
	return int32(c.builtinStringAddrs[code])
}

func (c *Compiler) local(l ir.Local) uint32 {
	var u32 uint32
	var exist bool
	if u32, exist = c.locals[l]; !exist {
		u32 = c.nextLocal
		c.locals[l] = u32
		c.nextLocal++
	}
	return u32
}

func (c *Compiler) genLocal() uint32 {
	l := c.nextLocal
	c.nextLocal++
	return l
}

func (c *Compiler) function(name string) uint32 {
	return c.funcs[name]
}

func (c *Compiler) appendInstr(instr instruction.Instruction) {
	c.code.Func.Expr.Instrs = append(c.code.Func.Expr.Instrs, instr)
}

func (c *Compiler) appendInstrs(instrs []instruction.Instruction) {
	for _, instr := range instrs {
		c.appendInstr(instr)
	}
}

func getLowestFreeDataSegmentOffset(m *module.Module) (int32, error) {

	var offset int32

	for i := range m.Data.Segments {

		if len(m.Data.Segments[i].Offset.Instrs) != 1 {
			return 0, errors.New("bad data segment offset instructions")
		}

		instr, ok := m.Data.Segments[i].Offset.Instrs[0].(instruction.I32Const)
		if !ok {
			return 0, errors.New("bad data segment offset expr")
		}

		// NOTE(tsandall): assume memory up to but not including addr is taken.
		addr := instr.Value + int32(len(m.Data.Segments[i].Init))
		if addr > offset {
			offset = addr
		}
	}

	return offset, nil
}
