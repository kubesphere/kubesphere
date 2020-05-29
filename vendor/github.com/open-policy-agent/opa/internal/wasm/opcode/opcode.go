// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package opcode contains constants and utilities for working with WASM opcodes.
package opcode

// Opcode represents a WASM instruction opcode.
type Opcode byte

// Control instructions.
const (
	Unreachable Opcode = iota
	Nop
	Block
	Loop
	If
	Else
)

const (
	// End defines the special end WASM opcode.
	End Opcode = 0x0B
)

// Extended control instructions.
const (
	Br Opcode = iota + 0x0C
	BrIf
	BrTable
	Return
	Call
	CallIndirect
)

// Parameter instructions.
const (
	Drop Opcode = iota + 0x1A
	Select
)

// Variable instructions.
const (
	GetLocal Opcode = iota + 0x20
	SetLocal
	TeeLocal
	GetGlobal
	SetGlobal
)

// Memory instructions.
const (
	I32Load Opcode = iota + 0x28
	I64Load
	F32Load
	F64Load
	I32Load8S
	I32Load8U
	I32Load16S
	I32Load16U
	I64Load8S
	I64Load8U
	I64Load16S
	I64Load16U
	I64Load32S
	I64Load32U
	I32Store
	I64Store
	F32Store
	F64Store
	I32Store8
	I32Store16
	I64Store8
	I64Store16
	I64Store32
	MemorySize
	MemoryGrow
)

// Numeric instructions.
const (
	I32Const Opcode = iota + 0x41
	I64Const
	F32Const
	F64Const

	I32Eqz
	I32Eq
	I32Ne
	I32LtS
	I32LtU
	I32GtS
	I32GtU
	I32LeS
	I32LeU
	I32GeS
	I32GeU

	I64Eqz
	I64Eq
	I64Ne
	I64LtS
	I64LtU
	I64GtS
	I64GtU
	I64LeS
	I64LeU
	I64GeS
	I64GeU

	F32Eq
	F32Ne
	F32Lt
	F32Gt
	F32Le
	F32Ge

	F64Eq
	F64Ne
	F64Lt
	F64Gt
	F64Le
	F64Ge

	I32Clz
	I32Ctz
	I32Popcnt
	I32Add
	I32Sub
	I32Mul
	I32DivS
	I32DivU
	I32RemS
	I32RemU
	I32And
	I32Or
	I32Xor
	I32Shl
	I32ShrS
	I32ShrU
	I32Rotl
	I32Rotr

	I64Clz
	I64Ctz
	I64Popcnt
	I64Add
	I64Sub
	I64Mul
	I64DivS
	I64DivU
	I64RemS
	I64RemU
	I64And
	I64Or
	I64Xor
	I64Shl
	I64ShrS
	I64ShrU
	I64Rotl
	I64Rotr

	F32Abs
	F32Neg
	F32Ceil
	F32Floor
	F32Trunc
	F32Nearest
	F32Sqrt
	F32Add
	F32Sub
	F32Mul
	F32Div
	F32Min
	F32Max
	F32Copysign

	F64Abs
	F64Neg
	F64Ceil
	F64Floor
	F64Trunc
	F64Nearest
	F64Sqrt
	F64Add
	F64Sub
	F64Mul
	F64Div
	F64Min
	F64Max
	F64Copysign

	I32WrapI64
	I32TruncSF32
	I32TruncUF32
	I32TruncSF64
	I32TruncUF64
	I64ExtendSI32
	I64ExtendUI32
	I64TruncSF32
	I64TruncUF32
	I64TruncSF64
	I64TruncUF64
	F32ConvertSI32
	F32ConvertUI32
	F32ConvertSI64
	F32ConvertUI64
	F32DemoteF64
	F64ConvertSI32
	F64ConvertUI32
	F64ConvertSI64
	F64ConvertUI64
	F64PromoteF32
	I32ReinterpretF32
	I64ReinterpretF64
	F32ReinterpretI32
	F64ReinterpretI64
)
