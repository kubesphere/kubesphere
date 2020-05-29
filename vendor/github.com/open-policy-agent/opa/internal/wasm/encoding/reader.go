// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/internal/leb128"
	"github.com/open-policy-agent/opa/internal/wasm/constant"
	"github.com/open-policy-agent/opa/internal/wasm/instruction"
	"github.com/open-policy-agent/opa/internal/wasm/module"
	"github.com/open-policy-agent/opa/internal/wasm/opcode"
	"github.com/open-policy-agent/opa/internal/wasm/types"
)

// ReadModule reads a binary-encoded WASM module from r.
func ReadModule(r io.Reader) (*module.Module, error) {

	wr := &reader{r: r, n: 0}
	module, err := readModule(wr)
	if err != nil {
		return nil, errors.Wrapf(err, "offset 0x%x", wr.n)
	}

	return module, nil
}

// ReadCodeEntry reads a binary-encoded WASM code entry from r.
func ReadCodeEntry(r io.Reader) (*module.CodeEntry, error) {

	wr := &reader{r: r, n: 0}
	entry, err := readCodeEntry(wr)
	if err != nil {
		return nil, errors.Wrapf(err, "offset 0x%x", wr.n)
	}

	return entry, nil
}

// CodeEntries returns the WASM code entries contained in r.
func CodeEntries(m *module.Module) ([]*module.CodeEntry, error) {

	entries := make([]*module.CodeEntry, len(m.Code.Segments))

	for i, s := range m.Code.Segments {
		buf := bytes.NewBuffer(s.Code)
		entry, err := ReadCodeEntry(buf)
		if err != nil {
			return nil, err
		}
		entries[i] = entry
	}

	return entries, nil
}

type reader struct {
	r io.Reader
	n int
}

func (r *reader) Read(bs []byte) (int, error) {
	n, err := r.r.Read(bs)
	r.n += n
	return n, err
}

func readModule(r io.Reader) (*module.Module, error) {

	if err := readMagic(r); err != nil {
		return nil, err
	}

	if err := readVersion(r); err != nil {
		return nil, err
	}

	var m module.Module

	if err := readSections(r, &m); err != nil && err != io.EOF {
		return nil, err
	}

	return &m, nil
}

func readCodeEntry(r io.Reader) (*module.CodeEntry, error) {

	var entry module.CodeEntry

	if err := readLocals(r, &entry.Func.Locals); err != nil {
		return nil, errors.Wrapf(err, "local declarations")
	}

	return &entry, readExpr(r, &entry.Func.Expr)
}

func readMagic(r io.Reader) error {
	var v uint32
	if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
		return err
	} else if v != constant.Magic {
		return fmt.Errorf("illegal magic value")
	}
	return nil
}

func readVersion(r io.Reader) error {
	var v uint32
	if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
		return err
	} else if v != constant.Version {
		return fmt.Errorf("illegal wasm version")
	}
	return nil
}

func readSections(r io.Reader, m *module.Module) error {
	for {
		id, err := readByte(r)
		if err != nil {
			return err
		}

		size, err := leb128.ReadVarUint32(r)
		if err != nil {
			return err
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}

		bufr := bytes.NewReader(buf)

		switch id {
		case constant.CustomSectionID, constant.StartSectionID, constant.MemorySectionID:
			continue
		case constant.TypeSectionID:
			if err := readTypeSection(bufr, &m.Type); err != nil {
				return errors.Wrap(err, "type section")
			}
		case constant.ImportSectionID:
			if err := readImportSection(bufr, &m.Import); err != nil {
				return errors.Wrap(err, "import section")
			}
		case constant.GlobalSectionID:
			if err := readGlobalSection(bufr, &m.Global); err != nil {
				return errors.Wrap(err, "global section")
			}
		case constant.TableSectionID:
			if err := readTableSection(bufr, &m.Table); err != nil {
				return errors.Wrap(err, "table section")
			}
		case constant.FunctionSectionID:
			if err := readFunctionSection(bufr, &m.Function); err != nil {
				return errors.Wrap(err, "function section")
			}
		case constant.ExportSectionID:
			if err := readExportSection(bufr, &m.Export); err != nil {
				return errors.Wrap(err, "export section")
			}
		case constant.ElementSectionID:
			if err := readElementSection(bufr, &m.Element); err != nil {
				return errors.Wrap(err, "element section")
			}
		case constant.DataSectionID:
			if err := readDataSection(bufr, &m.Data); err != nil {
				return errors.Wrap(err, "data section")
			}
		case constant.CodeSectionID:
			if err := readRawCodeSection(bufr, &m.Code); err != nil {
				return errors.Wrap(err, "code section")
			}
		default:
			return fmt.Errorf("illegal section id")
		}
	}
}

func readTypeSection(r io.Reader, s *module.TypeSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var ftype module.FunctionType
		if err := readFunctionType(r, &ftype); err != nil {
			return err
		}

		s.Functions = append(s.Functions, ftype)
	}

	return nil
}

func readImportSection(r io.Reader, s *module.ImportSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var imp module.Import

		if err := readImport(r, &imp); err != nil {
			return err
		}

		s.Imports = append(s.Imports, imp)
	}

	return nil
}

func readTableSection(r io.Reader, s *module.TableSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var table module.Table

		if elem, err := readByte(r); err != nil {
			return err
		} else if elem != constant.ElementTypeAnyFunc {
			return fmt.Errorf("illegal element type")
		} else {
			table.Type = types.Anyfunc
		}

		if err := readLimits(r, &table.Lim); err != nil {
			return err
		}

		s.Tables = append(s.Tables, table)
	}

	return nil
}

func readGlobalSection(r io.Reader, s *module.GlobalSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var global module.Global

		if err := readGlobal(r, &global); err != nil {
			return err
		}

		s.Globals = append(s.Globals, global)
	}

	return nil
}

func readFunctionSection(r io.Reader, s *module.FunctionSection) error {
	return readVarUint32Vector(r, &s.TypeIndices)
}

func readExportSection(r io.Reader, s *module.ExportSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var exp module.Export

		if err := readExport(r, &exp); err != nil {
			return err
		}

		s.Exports = append(s.Exports, exp)
	}

	return nil
}

func readElementSection(r io.Reader, s *module.ElementSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var seg module.ElementSegment

		if err := readElementSegment(r, &seg); err != nil {
			return err
		}

		s.Segments = append(s.Segments, seg)
	}

	return nil
}

func readDataSection(r io.Reader, s *module.DataSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {

		var seg module.DataSegment

		if err := readDataSegment(r, &seg); err != nil {
			return err
		}

		s.Segments = append(s.Segments, seg)
	}

	return nil
}

func readRawCodeSection(r io.Reader, s *module.RawCodeSection) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < n; i++ {
		var seg module.RawCodeSegment

		if err := readRawCodeSegment(r, &seg); err != nil {
			return err
		}

		s.Segments = append(s.Segments, seg)
	}

	return nil
}

func readFunctionType(r io.Reader, ftype *module.FunctionType) error {

	if b, err := readByte(r); err != nil {
		return err
	} else if b != constant.FunctionTypeID {
		return fmt.Errorf("illegal function type id 0x%x", b)
	}

	if err := readValueTypeVector(r, &ftype.Params); err != nil {
		return err
	}

	return readValueTypeVector(r, &ftype.Results)
}

func readGlobal(r io.Reader, global *module.Global) error {

	if err := readValueType(r, &global.Type); err != nil {
		return err
	}

	b, err := readByte(r)
	if err != nil {
		return err
	}

	if b == 1 {
		global.Mutable = true
	} else if b != 0 {
		return fmt.Errorf("illegal mutability flag")
	}

	if err := readConstantExpr(r, &global.Init); err != nil {
		return err
	}

	return nil
}

func readImport(r io.Reader, imp *module.Import) error {

	if err := readByteVectorString(r, &imp.Module); err != nil {
		return err
	}

	if err := readByteVectorString(r, &imp.Name); err != nil {
		return err
	}

	b, err := readByte(r)
	if err != nil {
		return err

	}

	if b == constant.ImportDescType {
		index, err := leb128.ReadVarUint32(r)
		if err != nil {
			return err
		}
		imp.Descriptor = module.FunctionImport{
			Func: index,
		}
		return nil
	}

	if b == constant.ImportDescTable {
		if elem, err := readByte(r); err != nil {
			return err
		} else if elem != constant.ElementTypeAnyFunc {
			return fmt.Errorf("illegal element type")
		}
		desc := module.TableImport{
			Type: types.Anyfunc,
		}
		if err := readLimits(r, &desc.Lim); err != nil {
			return err
		}
		imp.Descriptor = desc
		return nil
	}

	if b == constant.ImportDescMem {
		desc := module.MemoryImport{}
		if err := readLimits(r, &desc.Mem.Lim); err != nil {
			return err
		}
		imp.Descriptor = desc
		return nil
	}

	if b == constant.ImportDescGlobal {
		desc := module.GlobalImport{}
		if err := readValueType(r, &desc.Type); err != nil {
			return err
		}
		b, err := readByte(r)
		if err != nil {
			return err
		}
		if b == 1 {
			desc.Mutable = true
		} else if b != 0 {
			return fmt.Errorf("illegal mutability flag")
		}
		return nil
	}

	return fmt.Errorf("illegal import descriptor type")
}

func readExport(r io.Reader, exp *module.Export) error {

	if err := readByteVectorString(r, &exp.Name); err != nil {
		return err
	}

	b, err := readByte(r)
	if err != nil {
		return err
	}

	switch b {
	case constant.ExportDescType:
		exp.Descriptor.Type = module.FunctionExportType
	case constant.ExportDescTable:
		exp.Descriptor.Type = module.TableExportType
	case constant.ExportDescMem:
		exp.Descriptor.Type = module.MemoryExportType
	case constant.ExportDescGlobal:
		exp.Descriptor.Type = module.GlobalExportType
	default:
		return fmt.Errorf("illegal export descriptor type")
	}

	exp.Descriptor.Index, err = leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	return nil
}

func readElementSegment(r io.Reader, seg *module.ElementSegment) error {

	if err := readVarUint32(r, &seg.Index); err != nil {
		return err
	}

	if err := readConstantExpr(r, &seg.Offset); err != nil {
		return err
	}

	if err := readVarUint32Vector(r, &seg.Indices); err != nil {
		return err
	}

	return nil
}

func readDataSegment(r io.Reader, seg *module.DataSegment) error {

	if err := readVarUint32(r, &seg.Index); err != nil {
		return err
	}

	if err := readConstantExpr(r, &seg.Offset); err != nil {
		return err
	}

	if err := readByteVector(r, &seg.Init); err != nil {
		return err
	}

	return nil
}

func readRawCodeSegment(r io.Reader, seg *module.RawCodeSegment) error {
	return readByteVector(r, &seg.Code)
}

func readConstantExpr(r io.Reader, expr *module.Expr) error {

	instrs := make([]instruction.Instruction, 0)

	for {
		b, err := readByte(r)
		if err != nil {
			return err
		}

		switch opcode.Opcode(b) {
		case opcode.I32Const:
			i32, err := leb128.ReadVarInt32(r)
			if err != nil {
				return err
			}
			instrs = append(instrs, instruction.I32Const{Value: i32})
		case opcode.I64Const:
			i64, err := leb128.ReadVarInt64(r)
			if err != nil {
				return err
			}
			instrs = append(instrs, instruction.I64Const{Value: i64})
		case opcode.End:
			expr.Instrs = instrs
			return nil
		default:
			return fmt.Errorf("illegal constant expr opcode 0x%x", b)
		}
	}
}

func readExpr(r io.Reader, expr *module.Expr) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	return readInstructions(r, &expr.Instrs)
}

func readInstructions(r io.Reader, instrs *[]instruction.Instruction) error {

	ret := make([]instruction.Instruction, 0)

	for {
		b, err := readByte(r)
		if err != nil {
			return err
		}

		switch opcode.Opcode(b) {
		case opcode.I32Const:
			ret = append(ret, instruction.I32Const{Value: leb128.MustReadVarInt32(r)})
		case opcode.I64Const:
			ret = append(ret, instruction.I64Const{Value: leb128.MustReadVarInt64(r)})
		case opcode.I32Eqz:
			ret = append(ret, instruction.I32Eqz{})
		case opcode.GetLocal:
			ret = append(ret, instruction.GetLocal{Index: leb128.MustReadVarUint32(r)})
		case opcode.SetLocal:
			ret = append(ret, instruction.SetLocal{Index: leb128.MustReadVarUint32(r)})
		case opcode.Call:
			ret = append(ret, instruction.Call{Index: leb128.MustReadVarUint32(r)})
		case opcode.BrIf:
			ret = append(ret, instruction.BrIf{Index: leb128.MustReadVarUint32(r)})
		case opcode.Return:
			ret = append(ret, instruction.Return{})
		case opcode.Block:
			block := instruction.Block{}
			if err := readBlockValueType(r, block.Type); err != nil {
				return err
			}
			if err := readInstructions(r, &block.Instrs); err != nil {
				return err
			}
			ret = append(ret, block)
		case opcode.Loop:
			loop := instruction.Loop{}
			if err := readBlockValueType(r, loop.Type); err != nil {
				return err
			}
			if err := readInstructions(r, &loop.Instrs); err != nil {
				return err
			}
			ret = append(ret, loop)
		case opcode.End:
			*instrs = ret
			return nil
		default:
			return fmt.Errorf("illegal opcode 0x%x", b)
		}
	}
}

func readLimits(r io.Reader, l *module.Limit) error {

	b, err := readByte(r)
	if err != nil {
		return err
	}

	min, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	l.Min = min

	if b == 1 {
		max, err := leb128.ReadVarUint32(r)
		if err != nil {
			return err
		}
		l.Max = &max
	} else if b != 0 {
		return fmt.Errorf("illegal limit flag")
	}

	return nil
}

func readLocals(r io.Reader, locals *[]module.LocalDeclaration) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	ret := make([]module.LocalDeclaration, n)

	for i := uint32(0); i < n; i++ {
		if err := readVarUint32(r, &ret[i].Count); err != nil {
			return err
		}
		if err := readValueType(r, &ret[i].Type); err != nil {
			return err
		}
	}

	*locals = ret
	return nil
}

func readByteVector(r io.Reader, v *[]byte) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	*v = buf
	return nil
}

func readByteVectorString(r io.Reader, v *string) error {

	var buf []byte

	if err := readByteVector(r, &buf); err != nil {
		return err
	}

	*v = string(buf)
	return nil
}

func readVarUint32Vector(r io.Reader, v *[]uint32) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	ret := make([]uint32, n)

	for i := uint32(0); i < n; i++ {
		if err := readVarUint32(r, &ret[i]); err != nil {
			return err
		}
	}

	*v = ret
	return nil
}

func readValueTypeVector(r io.Reader, v *[]types.ValueType) error {

	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	ret := make([]types.ValueType, n)

	for i := uint32(0); i < n; i++ {
		if err := readValueType(r, &ret[i]); err != nil {
			return err
		}
	}

	*v = ret
	return nil
}

func readVarUint32(r io.Reader, v *uint32) error {
	var err error
	*v, err = leb128.ReadVarUint32(r)
	return err
}

func readValueType(r io.Reader, v *types.ValueType) error {
	if b, err := readByte(r); err != nil {
		return err
	} else if b == constant.ValueTypeI32 {
		*v = types.I32
	} else if b == constant.ValueTypeI64 {
		*v = types.I64
	} else if b == constant.ValueTypeF32 {
		*v = types.F32
	} else if b == constant.ValueTypeF64 {
		*v = types.F64
	} else {
		return fmt.Errorf("illegal value type: 0x%x", b)
	}
	return nil
}

func readBlockValueType(r io.Reader, v *types.ValueType) error {
	if b, err := readByte(r); err != nil {
		return err
	} else if b == constant.ValueTypeI32 {
		*v = types.I32
	} else if b == constant.ValueTypeI64 {
		*v = types.I64
	} else if b == constant.ValueTypeF32 {
		*v = types.F32
	} else if b == constant.ValueTypeF64 {
		*v = types.F64
	} else if b != constant.BlockTypeEmpty {
		return fmt.Errorf("illegal value type: 0x%x", b)
	}
	return nil
}

func readByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	return buf[0], err
}
