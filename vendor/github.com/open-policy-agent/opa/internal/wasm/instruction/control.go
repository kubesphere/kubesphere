// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package instruction

import (
	"github.com/open-policy-agent/opa/internal/wasm/opcode"
	"github.com/open-policy-agent/opa/internal/wasm/types"
)

// Unreachable reprsents an unreachable opcode.
type Unreachable struct {
	NoImmediateArgs
}

// Op returns the opcode of the instruction.
func (Unreachable) Op() opcode.Opcode {
	return opcode.Unreachable
}

// Nop represents a WASM no-op instruction.
type Nop struct {
	NoImmediateArgs
}

// Op returns the opcode of the instruction.
func (Nop) Op() opcode.Opcode {
	return opcode.Nop
}

// Block represents a WASM block instruction.
type Block struct {
	NoImmediateArgs
	Type   *types.ValueType
	Instrs []Instruction
}

// Op returns the opcode of the instruction
func (Block) Op() opcode.Opcode {
	return opcode.Block
}

// BlockType returns the type of the block's return value.
func (i Block) BlockType() *types.ValueType {
	return i.Type
}

// Instructions returns the instructions contained in the block.
func (i Block) Instructions() []Instruction {
	return i.Instrs
}

// Loop represents a WASM loop instruction.
type Loop struct {
	NoImmediateArgs
	Type   *types.ValueType
	Instrs []Instruction
}

// Op returns the opcode of the instruction.
func (Loop) Op() opcode.Opcode {
	return opcode.Loop
}

// BlockType returns the type of the loop's return value.
func (i Loop) BlockType() *types.ValueType {
	return i.Type
}

// Instructions represents the instructions contained in the loop.
func (i Loop) Instructions() []Instruction {
	return i.Instrs
}

// Br represents a WASM br instruction.
type Br struct {
	Index uint32
}

// Op returns the opcode of the instruction.
func (Br) Op() opcode.Opcode {
	return opcode.Br
}

// ImmediateArgs returns the block index to break to.
func (i Br) ImmediateArgs() []interface{} {
	return []interface{}{i.Index}
}

// BrIf represents a WASM br_if instruction.
type BrIf struct {
	Index uint32
}

// Op returns the opcode of the instruction.
func (BrIf) Op() opcode.Opcode {
	return opcode.BrIf
}

// ImmediateArgs returns the block index to break to.
func (i BrIf) ImmediateArgs() []interface{} {
	return []interface{}{i.Index}
}

// Call represents a WASM call instruction.
type Call struct {
	Index uint32
}

// Op returns the opcode of the instruction.
func (Call) Op() opcode.Opcode {
	return opcode.Call
}

// ImmediateArgs returns the function index.
func (i Call) ImmediateArgs() []interface{} {
	return []interface{}{i.Index}
}

// Return represents a WASM return instruction.
type Return struct {
	NoImmediateArgs
}

// Op returns the opcode of the instruction.
func (Return) Op() opcode.Opcode {
	return opcode.Return
}

// End represents the special WASM end instruction.
type End struct {
	NoImmediateArgs
}

// Op returns the opcode of the instruction.
func (End) Op() opcode.Opcode {
	return opcode.End
}
