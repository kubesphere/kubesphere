package csvutil

import (
	"io"
	"reflect"
)

type decField struct {
	columnIndex int
	field
	decodeFunc
	zero interface{}
}

// A Decoder reads and decodes string records into structs.
type Decoder struct {
	// Tag defines which key in the struct field's tag to scan for names and
	// options (Default: 'csv').
	Tag string

	// If true, Decoder will return a MissingColumnsError if it discovers
	// that any of the columns are missing. This means that a CSV input
	// will be required to contain all columns that were defined in the
	// provided struct.
	DisallowMissingColumns bool

	// If not nil, Map is a function that is called for each field in the csv
	// record before decoding the data. It allows mapping certain string values
	// for specific columns or types to a known format. Decoder calls Map with
	// the current column name (taken from header) and a zero non-pointer value
	// of a type to which it is going to decode data into. Implementations
	// should use type assertions to recognize the type.
	//
	// The good example of use case for Map is if NaN values are represented by
	// eg 'n/a' string, implementing a specific Map function for all floats
	// could map 'n/a' back into 'NaN' to allow successful decoding.
	//
	// Use Map with caution. If the requirements of column or type are not met
	// Map should return 'field', since it is the original value that was
	// read from the csv input, this would indicate no change.
	//
	// If struct field is an interface v will be of type string, unless the
	// struct field contains a settable pointer value - then v will be a zero
	// value of that type.
	//
	// Map must be set before the first call to Decode and not changed after it.
	Map func(field, col string, v interface{}) string

	r          Reader
	typeKey    typeKey
	hmap       map[string]int
	header     []string
	record     []string
	cache      []decField
	unused     []int
	funcMap    map[reflect.Type]reflect.Value
	ifaceFuncs []reflect.Value
}

// NewDecoder returns a new decoder that reads from r.
//
// Decoder will match struct fields according to the given header.
//
// If header is empty NewDecoder will read one line and treat it as a header.
//
// Records coming from r must be of the same length as the header.
//
// NewDecoder may return io.EOF if there is no data in r and no header was
// provided by the caller.
func NewDecoder(r Reader, header ...string) (dec *Decoder, err error) {
	if len(header) == 0 {
		header, err = r.Read()
		if err != nil {
			return nil, err
		}
	}

	h := make([]string, len(header))
	copy(h, header)
	header = h

	m := make(map[string]int, len(header))
	for i, h := range header {
		m[h] = i
	}

	return &Decoder{
		r:      r,
		header: header,
		hmap:   m,
		unused: make([]int, 0, len(header)),
	}, nil
}

// Decode reads the next string record or records from its input and stores it
// in the value pointed to by v which must be a pointer to a struct, struct slice
// or struct array.
//
// Decode matches all exported struct fields based on the header. Struct fields
// can be adjusted by using tags.
//
// The "omitempty" option specifies that the field should be omitted from
// the decoding if record's field is an empty string.
//
// Examples of struct field tags and their meanings:
// 	// Decode matches this field with "myName" header column.
// 	Field int `csv:"myName"`
//
// 	// Decode matches this field with "Field" header column.
// 	Field int
//
// 	// Decode matches this field with "myName" header column and decoding is not
//	// called if record's field is an empty string.
// 	Field int `csv:"myName,omitempty"`
//
// 	// Decode matches this field with "Field" header column and decoding is not
//	// called if record's field is an empty string.
// 	Field int `csv:",omitempty"`
//
// 	// Decode ignores this field.
// 	Field int `csv:"-"`
//
//	// Decode treats this field exactly as if it was an embedded field and
//	// matches header columns that start with "my_prefix_" to all fields of this
//	// type.
//	Field Struct `csv:"my_prefix_,inline"`
//
//	// Decode treats this field exactly as if it was an embedded field.
//	Field Struct `csv:",inline"`
//
// By default decode looks for "csv" tag, but this can be changed by setting
// Decoder.Tag field.
//
// To Decode into a custom type v must implement csvutil.Unmarshaler or
// encoding.TextUnmarshaler.
//
// Anonymous struct fields with tags are treated like normal fields and they
// must implement csvutil.Unmarshaler or encoding.TextUnmarshaler unless inline
// tag is specified.
//
// Anonymous struct fields without tags are populated just as if they were
// part of the main struct. However, fields in the main struct have bigger
// priority and they are populated first. If main struct and anonymous struct
// field have the same fields, the main struct's fields will be populated.
//
// Fields of type []byte expect the data to be base64 encoded strings.
//
// Float fields are decoded to NaN if a string value is 'NaN'. This check
// is case insensitive.
//
// Interface fields are decoded to strings unless they contain settable pointer
// value.
//
// Pointer fields are decoded to nil if a string value is empty.
//
// If v is a slice, Decode resets it and reads the input until EOF, storing all
// decoded values in the given slice. Decode returns nil on EOF.
//
// If v is an array, Decode reads the input until EOF or until it decodes all
// corresponding array elements. If the input contains less elements than the
// array, the additional Go array elements are set to zero values. Decode
// returns nil on EOF unless there were no records decoded.
//
// Fields with inline tags that have a non-empty prefix must not be cyclic
// structures. Passing such values to Decode will result in an infinite loop.
func (d *Decoder) Decode(v interface{}) (err error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &InvalidDecodeError{Type: reflect.TypeOf(v)}
	}

	elem := indirect(val.Elem())
	switch elem.Kind() {
	case reflect.Struct:
		return d.decodeStruct(elem)
	case reflect.Slice:
		return d.decodeSlice(elem)
	case reflect.Array:
		return d.decodeArray(elem)
	case reflect.Interface, reflect.Invalid:
		elem = walkValue(elem)
		if elem.Kind() != reflect.Invalid {
			return &InvalidDecodeError{Type: elem.Type()}
		}
		return &InvalidDecodeError{Type: val.Type()}
	default:
		return &InvalidDecodeError{Type: reflect.PtrTo(elem.Type())}
	}
}

// Record returns the most recently read record. The slice is valid until the
// next call to Decode.
func (d *Decoder) Record() []string {
	return d.record
}

// Header returns the first line that came from the reader, or returns the
// defined header by the caller.
func (d *Decoder) Header() []string {
	header := make([]string, len(d.header))
	copy(header, d.header)
	return header
}

// Unused returns a list of column indexes that were not used during decoding
// due to lack of matching struct field.
func (d *Decoder) Unused() []int {
	if len(d.unused) == 0 {
		return nil
	}

	indices := make([]int, len(d.unused))
	copy(indices, d.unused)
	return indices
}

// Register registers a custom decoding function for a concrete type or interface.
// The argument f must be of type:
// 	func([]byte, T) error
//
// T must be a concrete type such as *time.Time, or interface that has at least one
// method.
//
// During decoding, fields are matched by the concrete type first. If match is not
// found then Decoder looks if field implements any of the registered interfaces
// in order they were registered.
//
// Register panics if:
//	- f does not match the right signature
//	- f is an empty interface
//	- f was already registered
//
// Register is based on the encoding/json proposal:
// https://github.com/golang/go/issues/5901.
func (d *Decoder) Register(f interface{}) {
	v := reflect.ValueOf(f)
	typ := v.Type()

	if typ.Kind() != reflect.Func ||
		typ.NumIn() != 2 || typ.NumOut() != 1 ||
		typ.In(0) != _bytes || typ.Out(0) != _error {
		panic("csvutil: func must be of type func([]byte, T) error")
	}

	argType := typ.In(1)

	if argType.Kind() == reflect.Interface && argType.NumMethod() == 0 {
		panic("csvutil: func argument type must not be an empty interface")
	}

	if d.funcMap == nil {
		d.funcMap = make(map[reflect.Type]reflect.Value)
	}

	if _, ok := d.funcMap[argType]; ok {
		panic("csvutil: func " + typ.String() + " already registered")
	}

	d.funcMap[argType] = v

	if argType.Kind() == reflect.Interface {
		d.ifaceFuncs = append(d.ifaceFuncs, v)
	}
}

func (d *Decoder) decodeSlice(slice reflect.Value) error {
	typ := slice.Type().Elem()
	if walkType(typ).Kind() != reflect.Struct {
		return &InvalidDecodeError{Type: reflect.PtrTo(slice.Type())}
	}

	slice.SetLen(0)

	var c int
	for ; ; c++ {
		v := reflect.New(typ)

		err := d.decodeStruct(indirect(v))
		if err == io.EOF {
			if c == 0 {
				return io.EOF
			}
			break
		}

		// we want to ensure that we append this element to the slice even if it
		// was partially decoded due to error. This is how JSON pkg does it.
		slice.Set(reflect.Append(slice, v.Elem()))
		if err != nil {
			return err
		}
	}

	slice.Set(slice.Slice3(0, c, c))
	return nil
}

func (d *Decoder) decodeArray(v reflect.Value) error {
	if walkType(v.Type().Elem()).Kind() != reflect.Struct {
		return &InvalidDecodeError{Type: reflect.PtrTo(v.Type())}
	}

	l := v.Len()

	var i int
	for ; i < l; i++ {
		if err := d.decodeStruct(indirect(v.Index(i))); err == io.EOF {
			if i == 0 {
				return io.EOF
			}
			break
		} else if err != nil {
			return err
		}
	}

	zero := reflect.Zero(v.Type().Elem())
	for i := i; i < l; i++ {
		v.Index(i).Set(zero)
	}
	return nil
}

func (d *Decoder) decodeStruct(v reflect.Value) (err error) {
	d.record, err = d.r.Read()
	if err != nil {
		return err
	}

	if len(d.record) != len(d.header) {
		return ErrFieldCount
	}

	return d.unmarshal(d.record, v)
}

func (d *Decoder) unmarshal(record []string, v reflect.Value) error {
	fields, err := d.fields(typeKey{d.tag(), v.Type()})
	if err != nil {
		return err
	}

fieldLoop:
	for _, f := range fields {
		isBlank := record[f.columnIndex] == ""
		if f.tag.omitEmpty && isBlank {
			continue
		}

		fv := v
		for n, i := range f.index {
			fv = fv.Field(i)
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					if isBlank && n == len(f.index)-1 { // ensure we are on the leaf.
						continue fieldLoop
					}
					// this can happen if a field is an unexported embedded
					// pointer type. In Go prior to 1.10 it was possible to
					// set such value because of a bug in the reflect package
					// https://github.com/golang/go/issues/21353
					if !fv.CanSet() {
						return errPtrUnexportedStruct(fv.Type())
					}
					fv.Set(reflect.New(fv.Type().Elem()))
				}

				if isBlank && n == len(f.index)-1 { // ensure we are on the leaf.
					fv.Set(reflect.Zero(fv.Type()))
					continue fieldLoop
				}

				if n != len(f.index)-1 {
					fv = fv.Elem() // walk pointer until we are on the the leaf.
				}
			}
		}

		s := record[f.columnIndex]
		if d.Map != nil && f.zero != nil {
			zero := f.zero
			if fv := walkPtr(fv); fv.Kind() == reflect.Interface && !fv.IsNil() {
				if v := walkValue(fv); v.CanSet() {
					zero = reflect.Zero(v.Type()).Interface()
				}
			}
			s = d.Map(s, d.header[f.columnIndex], zero)
		}

		if err := f.decodeFunc(s, fv); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) fields(k typeKey) ([]decField, error) {
	if k == d.typeKey {
		return d.cache, nil
	}

	var (
		fields      = cachedFields(k)
		decFields   = make([]decField, 0, len(fields))
		used        = make([]bool, len(d.header))
		missingCols []string
	)
	for _, f := range fields {
		i, ok := d.hmap[f.name]
		if !ok {
			if d.DisallowMissingColumns {
				missingCols = append(missingCols, f.name)
			}
			continue
		}

		fn, err := decodeFn(f.baseType, d.funcMap, d.ifaceFuncs)
		if err != nil {
			return nil, err
		}

		df := decField{
			columnIndex: i,
			field:       f,
			decodeFunc:  fn,
		}

		if d.Map != nil {
			switch f.typ.Kind() {
			case reflect.Interface:
				df.zero = "" // interface values are decoded to strings
			default:
				df.zero = reflect.Zero(walkType(f.typ)).Interface()
			}
		}

		decFields = append(decFields, df)
		used[i] = true
	}

	if len(missingCols) > 0 {
		return nil, &MissingColumnsError{
			Columns: missingCols,
		}
	}

	d.unused = d.unused[:0]
	for i, b := range used {
		if !b {
			d.unused = append(d.unused, i)
		}
	}

	d.cache, d.typeKey = decFields, k
	return d.cache, nil
}

func (d *Decoder) tag() string {
	if d.Tag == "" {
		return defaultTag
	}
	return d.Tag
}

func indirect(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
			return v
		case reflect.Ptr:
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		default:
			return v
		}
	}
}
