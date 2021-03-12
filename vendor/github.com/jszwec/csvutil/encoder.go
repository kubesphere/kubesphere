package csvutil

import (
	"reflect"
)

const defaultBufSize = 4096

type encField struct {
	field
	encodeFunc
}

type encCache struct {
	fields []encField
	buf    []byte
	index  []int
	record []string
}

func newEncCache(k typeKey, funcMap map[reflect.Type]reflect.Value, funcs []reflect.Value) (_ *encCache, err error) {
	fields := cachedFields(k)
	encFields := make([]encField, len(fields))

	for i, f := range fields {
		fn, err := encodeFn(f.baseType, true, funcMap, funcs)
		if err != nil {
			return nil, err
		}

		encFields[i] = encField{
			field:      f,
			encodeFunc: fn,
		}
	}
	return &encCache{
		fields: encFields,
		buf:    make([]byte, 0, defaultBufSize),
		index:  make([]int, len(encFields)),
		record: make([]string, len(encFields)),
	}, nil
}

// Encoder writes structs CSV representations to the output stream.
type Encoder struct {
	// Tag defines which key in the struct field's tag to scan for names and
	// options (Default: 'csv').
	Tag string

	// If AutoHeader is true, a struct header is encoded during the first call
	// to Encode automatically (Default: true).
	AutoHeader bool

	w          Writer
	c          *encCache
	noHeader   bool
	typeKey    typeKey
	funcMap    map[reflect.Type]reflect.Value
	ifaceFuncs []reflect.Value
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w Writer) *Encoder {
	return &Encoder{
		w:          w,
		noHeader:   true,
		AutoHeader: true,
	}
}

// Register registers a custom encoding function for a concrete type or interface.
// The argument f must be of type:
// 	func(T) ([]byte, error)
//
// T must be a concrete type such as Foo or *Foo, or interface that has at
// least one method.
//
// During encoding, fields are matched by the concrete type first. If match is not
// found then Encoder looks if field implements any of the registered interfaces
// in order they were registered.
//
// Register panics if:
//	- f does not match the right signature
//	- f is an empty interface
//	- f was already registered
//
// Register is based on the encoding/json proposal:
// https://github.com/golang/go/issues/5901.
func (e *Encoder) Register(f interface{}) {
	v := reflect.ValueOf(f)
	typ := v.Type()

	if typ.Kind() != reflect.Func ||
		typ.NumIn() != 1 || typ.NumOut() != 2 ||
		typ.Out(0) != _bytes || typ.Out(1) != _error {
		panic("csvutil: func must be of type func(T) ([]byte, error)")
	}

	argType := typ.In(0)

	if argType.Kind() == reflect.Interface && argType.NumMethod() == 0 {
		panic("csvutil: func argument type must not be an empty interface")
	}

	if e.funcMap == nil {
		e.funcMap = make(map[reflect.Type]reflect.Value)
	}

	if _, ok := e.funcMap[argType]; ok {
		panic("csvutil: func " + typ.String() + " already registered")
	}

	e.funcMap[argType] = v

	if argType.Kind() == reflect.Interface {
		e.ifaceFuncs = append(e.ifaceFuncs, v)
	}
}

// Encode writes the CSV encoding of v to the output stream. The provided
// argument v must be a struct, struct slice or struct array.
//
// Only the exported fields will be encoded.
//
// First call to Encode will write a header unless EncodeHeader was called first
// or AutoHeader is false. Header names can be customized by using tags
// ('csv' by default), otherwise original Field names are used.
//
// Header and fields are written in the same order as struct fields are defined.
// Embedded struct's fields are treated as if they were part of the outer struct.
// Fields that are embedded types and that are tagged are treated like any
// other field, but they have to implement Marshaler or encoding.TextMarshaler
// interfaces.
//
// Marshaler interface has the priority over encoding.TextMarshaler.
//
// Tagged fields have the priority over non tagged fields with the same name.
//
// Following the Go visibility rules if there are multiple fields with the same
// name (tagged or not tagged) on the same level and choice between them is
// ambiguous, then all these fields will be ignored.
//
// Nil values will be encoded as empty strings. Same will happen if 'omitempty'
// tag is set, and the value is a default value like 0, false or nil interface.
//
// Bool types are encoded as 'true' or 'false'.
//
// Float types are encoded using strconv.FormatFloat with precision -1 and 'G'
// format. NaN values are encoded as 'NaN' string.
//
// Fields of type []byte are being encoded as base64-encoded strings.
//
// Fields can be excluded from encoding by using '-' tag option.
//
// Examples of struct tags:
//
// 	// Field appears as 'myName' header in CSV encoding.
// 	Field int `csv:"myName"`
//
// 	// Field appears as 'Field' header in CSV encoding.
// 	Field int
//
// 	// Field appears as 'myName' header in CSV encoding and is an empty string
//	// if Field is 0.
// 	Field int `csv:"myName,omitempty"`
//
// 	// Field appears as 'Field' header in CSV encoding and is an empty string
//	// if Field is 0.
// 	Field int `csv:",omitempty"`
//
// 	// Encode ignores this field.
// 	Field int `csv:"-"`
//
//	// Encode treats this field exactly as if it was an embedded field and adds
//	// "my_prefix_" to each field's name.
//	Field Struct `csv:"my_prefix_,inline"`
//
//	// Encode treats this field exactly as if it was an embedded field.
//	Field Struct `csv:",inline"`
//
// Fields with inline tags that have a non-empty prefix must not be cyclic
// structures. Passing such values to Encode will result in an infinite loop.
//
// Encode doesn't flush data. The caller is responsible for calling Flush() if
// the used Writer supports it.
func (e *Encoder) Encode(v interface{}) error {
	return e.encode(reflect.ValueOf(v))
}

// EncodeHeader writes the CSV header of the provided struct value to the output
// stream. The provided argument v must be a struct value.
//
// The first Encode method call will not write header if EncodeHeader was called
// before it. This method can be called in cases when a data set could be
// empty, but header is desired.
//
// EncodeHeader is like Header function, but it works with the Encoder and writes
// directly to the output stream. Look at Header documentation for the exact
// header encoding rules.
func (e *Encoder) EncodeHeader(v interface{}) error {
	typ, err := valueType(v)
	if err != nil {
		return err
	}
	return e.encodeHeader(typ)
}

func (e *Encoder) encode(v reflect.Value) error {
	val := walkValue(v)

	if !val.IsValid() {
		return &InvalidEncodeError{}
	}

	switch val.Kind() {
	case reflect.Struct:
		return e.encodeStruct(val)
	case reflect.Array, reflect.Slice:
		if walkType(val.Type().Elem()).Kind() != reflect.Struct {
			return &InvalidEncodeError{v.Type()}
		}
		return e.encodeArray(val)
	default:
		return &InvalidEncodeError{v.Type()}
	}
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	if e.AutoHeader && e.noHeader {
		if err := e.encodeHeader(v.Type()); err != nil {
			return err
		}
	}
	return e.marshal(v)
}

func (e *Encoder) encodeArray(v reflect.Value) error {
	l := v.Len()
	for i := 0; i < l; i++ {
		if err := e.encodeStruct(walkValue(v.Index(i))); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeHeader(typ reflect.Type) error {
	fields, _, _, record, err := e.cache(typ)
	if err != nil {
		return err
	}

	for i, f := range fields {
		record[i] = f.name
	}

	if err := e.w.Write(record); err != nil {
		return err
	}

	e.noHeader = false
	return nil
}

func (e *Encoder) marshal(v reflect.Value) error {
	fields, buf, index, record, err := e.cache(v.Type())
	if err != nil {
		return err
	}

	for i, f := range fields {
		v := walkIndex(v, f.index)

		omitempty := f.tag.omitEmpty
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			// We should disable omitempty for pointer and interface values,
			// because if it's nil we will automatically encode it as an empty
			// string. However, the initialized pointer should not be affected,
			// even if it's a default value.
			omitempty = false
		}

		if !v.IsValid() {
			index[i] = 0
			continue
		}

		b, err := f.encodeFunc(buf, v, omitempty)
		if err != nil {
			return err
		}
		index[i], buf = len(b)-len(buf), b
	}

	out := string(buf)
	for i, n := range index {
		record[i], out = out[:n], out[n:]
	}
	e.c.buf = buf[:0]

	return e.w.Write(record)
}

func (e *Encoder) tag() string {
	if e.Tag == "" {
		return defaultTag
	}
	return e.Tag
}

func (e *Encoder) cache(typ reflect.Type) ([]encField, []byte, []int, []string, error) {
	if k := (typeKey{e.tag(), typ}); k != e.typeKey {
		c, err := newEncCache(k, e.funcMap, e.ifaceFuncs)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		e.c, e.typeKey = c, k
	}
	return e.c.fields, e.c.buf[:0], e.c.index, e.c.record, nil
}

func walkIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		v = walkPtr(v)
		if !v.IsValid() {
			return reflect.Value{}
		}
		v = v.Field(i)
	}
	return v
}

func walkPtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func walkValue(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			v = v.Elem()
		default:
			return v
		}
	}
}

func walkType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}
