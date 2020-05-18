// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/open-policy-agent/opa/internal/leb128"
	"github.com/open-policy-agent/opa/internal/wasm/constant"
	"github.com/open-policy-agent/opa/internal/wasm/instruction"
	"github.com/open-policy-agent/opa/internal/wasm/module"
	"github.com/open-policy-agent/opa/internal/wasm/opcode"
	"github.com/open-policy-agent/opa/internal/wasm/types"
)

// WriteModule writes a binary-encoded representation of module to w.
func WriteModule(w io.Writer, module *module.Module) error {

	if err := writeMagic(w); err != nil {
		return err
	}

	if err := writeVersion(w); err != nil {
		return err
	}

	if module == nil {
		return nil
	}

	if err := writeTypeSection(w, module.Type); err != nil {
		return err
	}

	if err := writeImportSection(w, module.Import); err != nil {
		return err
	}

	if err := writeFunctionSection(w, module.Function); err != nil {
		return err
	}

	if err := writeTableSection(w, module.Table); err != nil {
		return err
	}

	if err := writeGlobalSection(w, module.Global); err != nil {
		return err
	}

	if err := writeExportSection(w, module.Export); err != nil {
		return err
	}

	if err := writeElementSection(w, module.Element); err != nil {
		return err
	}

	if err := writeRawCodeSection(w, module.Code); err != nil {
		return err
	}

	if err := writeDataSection(w, module.Data); err != nil {
		return err
	}

	return nil
}

// WriteCodeEntry writes a binary encoded representation of entry to w.
func WriteCodeEntry(w io.Writer, entry *module.CodeEntry) error {

	if err := leb128.WriteVarUint32(w, uint32(len(entry.Func.Locals))); err != nil {
		return err
	}

	for _, local := range entry.Func.Locals {

		if err := leb128.WriteVarUint32(w, local.Count); err != nil {
			return err
		}

		if err := writeValueType(w, local.Type); err != nil {
			return err
		}
	}

	return writeInstructions(w, entry.Func.Expr.Instrs)
}

func writeMagic(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, constant.Magic)
}

func writeVersion(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, constant.Version)
}

func writeTypeSection(w io.Writer, s module.TypeSection) error {

	if len(s.Functions) == 0 {
		return nil
	}

	if err := writeByte(w, constant.TypeSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Functions))); err != nil {
		return err
	}

	for _, fsig := range s.Functions {
		if err := writeFunctionType(&buf, fsig); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeImportSection(w io.Writer, s module.ImportSection) error {

	if len(s.Imports) == 0 {
		return nil
	}

	if err := writeByte(w, constant.ImportSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Imports))); err != nil {
		return err
	}

	for _, imp := range s.Imports {
		if err := writeImport(&buf, imp); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeGlobalSection(w io.Writer, s module.GlobalSection) error {

	if len(s.Globals) == 0 {
		return nil
	}

	if err := writeByte(w, constant.GlobalSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Globals))); err != nil {
		return err
	}

	for _, global := range s.Globals {
		if err := writeGlobal(&buf, global); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeFunctionSection(w io.Writer, s module.FunctionSection) error {

	if len(s.TypeIndices) == 0 {
		return nil
	}

	if err := writeByte(w, constant.FunctionSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.TypeIndices))); err != nil {
		return err
	}

	for _, idx := range s.TypeIndices {
		if err := leb128.WriteVarUint32(&buf, uint32(idx)); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeTableSection(w io.Writer, s module.TableSection) error {

	if len(s.Tables) == 0 {
		return nil
	}

	if err := writeByte(w, constant.TableSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Tables))); err != nil {
		return err
	}

	for _, table := range s.Tables {
		switch table.Type {
		case types.Anyfunc:
			if err := writeByte(&buf, constant.ElementTypeAnyFunc); err != nil {
				return err
			}
		default:
			return fmt.Errorf("illegal table element type")
		}
		if err := writeLimits(&buf, table.Lim); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)

}

func writeExportSection(w io.Writer, s module.ExportSection) error {

	if len(s.Exports) == 0 {
		return nil
	}

	if err := writeByte(w, constant.ExportSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Exports))); err != nil {
		return err
	}

	for _, exp := range s.Exports {
		if err := writeByteVector(&buf, []byte(exp.Name)); err != nil {
			return err
		}
		var tpe byte
		switch exp.Descriptor.Type {
		case module.FunctionExportType:
			tpe = constant.ExportDescType
		case module.TableExportType:
			tpe = constant.ExportDescTable
		case module.MemoryExportType:
			tpe = constant.ExportDescMem
		case module.GlobalExportType:
			tpe = constant.ExportDescGlobal
		default:
			return fmt.Errorf("illegal export descriptor type 0x%x", exp.Descriptor.Type)
		}
		if err := writeByte(&buf, tpe); err != nil {
			return err
		}
		if err := leb128.WriteVarUint32(&buf, exp.Descriptor.Index); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeElementSection(w io.Writer, s module.ElementSection) error {

	if len(s.Segments) == 0 {
		return nil
	}

	if err := writeByte(w, constant.ElementSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Segments))); err != nil {
		return err
	}

	for _, seg := range s.Segments {
		if err := leb128.WriteVarUint32(&buf, seg.Index); err != nil {
			return err
		}
		if err := writeInstructions(&buf, seg.Offset.Instrs); err != nil {
			return err
		}
		if err := writeVarUint32Vector(&buf, seg.Indices); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeRawCodeSection(w io.Writer, s module.RawCodeSection) error {

	if len(s.Segments) == 0 {
		return nil
	}

	if err := writeByte(w, constant.CodeSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Segments))); err != nil {
		return err
	}

	for _, seg := range s.Segments {
		if err := leb128.WriteVarUint32(&buf, uint32(len(seg.Code))); err != nil {
			return err
		}
		if _, err := buf.Write(seg.Code); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeDataSection(w io.Writer, s module.DataSection) error {

	if len(s.Segments) == 0 {
		return nil
	}

	if err := writeByte(w, constant.DataSectionID); err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := leb128.WriteVarUint32(&buf, uint32(len(s.Segments))); err != nil {
		return err
	}

	for _, seg := range s.Segments {
		if err := leb128.WriteVarUint32(&buf, seg.Index); err != nil {
			return err
		}
		if err := writeInstructions(&buf, seg.Offset.Instrs); err != nil {
			return err
		}
		if err := writeByteVector(&buf, seg.Init); err != nil {
			return err
		}
	}

	return writeRawSection(w, &buf)
}

func writeFunctionType(w io.Writer, fsig module.FunctionType) error {

	if err := writeByte(w, constant.FunctionTypeID); err != nil {
		return err
	}

	if err := writeValueTypeVector(w, fsig.Params); err != nil {
		return err
	}

	return writeValueTypeVector(w, fsig.Results)
}

func writeImport(w io.Writer, imp module.Import) error {

	if err := writeByteVector(w, []byte(imp.Module)); err != nil {
		return err
	}

	if err := writeByteVector(w, []byte(imp.Name)); err != nil {
		return err
	}

	switch desc := imp.Descriptor.(type) {
	case module.FunctionImport:
		if err := writeByte(w, constant.ImportDescType); err != nil {
			return err
		}
		return leb128.WriteVarUint32(w, desc.Func)
	case module.TableImport:
		if err := writeByte(w, constant.ImportDescTable); err != nil {
			return err
		}
		if err := writeByte(w, constant.ElementTypeAnyFunc); err != nil {
			return err
		}
		return writeLimits(w, desc.Lim)
	case module.MemoryImport:
		if err := writeByte(w, constant.ImportDescMem); err != nil {
			return err
		}
		return writeLimits(w, desc.Mem.Lim)
	case module.GlobalImport:
		if err := writeByte(w, constant.ImportDescGlobal); err != nil {
			return err
		}
		if err := writeValueType(w, desc.Type); err != nil {
			return err
		}
		if desc.Mutable {
			return writeByte(w, constant.Mutable)
		}
		return writeByte(w, constant.Const)
	default:
		return fmt.Errorf("illegal import descriptor type")
	}
}

func writeGlobal(w io.Writer, global module.Global) error {

	if err := writeValueType(w, global.Type); err != nil {
		return err
	}

	var err error

	if global.Mutable {
		err = writeByte(w, constant.Mutable)
	} else {
		err = writeByte(w, constant.Const)
	}

	if err != nil {
		return err
	}

	if err := writeInstructions(w, global.Init.Instrs); err != nil {
		return err
	}

	return nil
}

func writeInstructions(w io.Writer, instrs []instruction.Instruction) error {

	for i, instr := range instrs {

		_, err := w.Write([]byte{byte(instr.Op())})
		if err != nil {
			return err
		}

		for _, arg := range instr.ImmediateArgs() {
			var err error
			switch arg := arg.(type) {
			case int32:
				err = leb128.WriteVarInt32(w, arg)
			case int64:
				err = leb128.WriteVarInt64(w, arg)
			case uint32:
				err = leb128.WriteVarUint32(w, arg)
			case uint64:
				err = leb128.WriteVarUint64(w, arg)
			case float32:
				u32 := math.Float32bits(arg)
				err = binary.Write(w, binary.LittleEndian, u32)
			case float64:
				u64 := math.Float64bits(arg)
				err = binary.Write(w, binary.LittleEndian, u64)
			default:
				return fmt.Errorf("illegal immediate argument type on instruction %d", i)
			}
			if err != nil {
				return err
			}
		}

		if si, ok := instr.(instruction.StructuredInstruction); ok {
			if err := writeBlockValueType(w, si.BlockType()); err != nil {
				return err
			}
			if err := writeInstructions(w, si.Instructions()); err != nil {
				return err
			}
		}

	}

	_, err := w.Write([]byte{byte(opcode.End)})
	return err
}

func writeLimits(w io.Writer, lim module.Limit) error {
	if lim.Max == nil {
		if err := writeByte(w, 0); err != nil {
			return err
		}
	} else {
		if err := writeByte(w, 1); err != nil {
			return err
		}
	}
	if err := leb128.WriteVarUint32(w, lim.Min); err != nil {
		return err
	}
	if lim.Max != nil {
		return leb128.WriteVarUint32(w, *lim.Max)
	}
	return nil
}

func writeVarUint32Vector(w io.Writer, v []uint32) error {

	if err := leb128.WriteVarUint32(w, uint32(len(v))); err != nil {
		return err
	}

	for i := range v {
		if err := leb128.WriteVarUint32(w, v[i]); err != nil {
			return err
		}
	}

	return nil
}

func writeValueTypeVector(w io.Writer, v []types.ValueType) error {

	if err := leb128.WriteVarUint32(w, uint32(len(v))); err != nil {
		return err
	}

	for i := range v {
		if err := writeValueType(w, v[i]); err != nil {
			return err
		}
	}

	return nil
}

func writeBlockValueType(w io.Writer, v *types.ValueType) error {
	var b byte
	if v != nil {
		switch *v {
		case types.I32:
			b = constant.ValueTypeI32
		case types.I64:
			b = constant.ValueTypeI64
		case types.F32:
			b = constant.ValueTypeF32
		case types.F64:
			b = constant.ValueTypeF64
		}
	} else {
		b = constant.BlockTypeEmpty
	}
	return writeByte(w, b)
}

func writeValueType(w io.Writer, v types.ValueType) error {
	var b byte
	switch v {
	case types.I32:
		b = constant.ValueTypeI32
	case types.I64:
		b = constant.ValueTypeI64
	case types.F32:
		b = constant.ValueTypeF32
	case types.F64:
		b = constant.ValueTypeF64
	}
	return writeByte(w, b)
}

func writeRawSection(w io.Writer, buf *bytes.Buffer) error {

	size := buf.Len()

	if err := leb128.WriteVarUint32(w, uint32(size)); err != nil {
		return err
	}

	_, err := io.Copy(w, buf)
	return err
}

func writeByteVector(w io.Writer, bs []byte) error {

	if err := leb128.WriteVarUint32(w, uint32(len(bs))); err != nil {
		return err
	}

	_, err := w.Write(bs)
	return err
}

func writeByte(w io.Writer, b byte) error {
	buf := make([]byte, 1)
	buf[0] = b
	_, err := w.Write(buf)
	return err
}
