package dbr

import (
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type interpolator struct {
	Buffer
	Dialect
	IgnoreBinary bool
	N            int
}

// InterpolateForDialect replaces placeholder in query with corresponding value in dialect
func InterpolateForDialect(query string, value []interface{}, d Dialect) (string, error) {
	i := interpolator{
		Buffer:  NewBuffer(),
		Dialect: d,
	}
	err := i.interpolate(query, value)
	if err != nil {
		return "", err
	}
	return i.String(), nil
}

func (i *interpolator) interpolate(query string, value []interface{}) error {
	if strings.Count(query, placeholder) != len(value) {
		return ErrPlaceholderCount
	}

	valueIndex := 0

	for {
		index := strings.Index(query, placeholder)
		if index == -1 {
			break
		}

		i.WriteString(query[:index])
		if _, ok := value[valueIndex].([]byte); ok && i.IgnoreBinary {
			i.WriteString(i.Placeholder(i.N))
			i.N++
			i.WriteValue(value[valueIndex])
		} else {
			err := i.encodePlaceholder(value[valueIndex])
			if err != nil {
				return err
			}
		}
		query = query[index+len(placeholder):]
		valueIndex++
	}

	// placeholder not found; write remaining query
	i.WriteString(query)

	return nil
}

func (i *interpolator) encodePlaceholder(value interface{}) error {
	if builder, ok := value.(Builder); ok {
		pbuf := NewBuffer()
		err := builder.Build(i.Dialect, pbuf)
		if err != nil {
			return err
		}
		paren := true
		switch value.(type) {
		case *SelectStmt:
		case *union:
		default:
			paren = false
		}
		if paren {
			i.WriteString("(")
		}
		err = i.interpolate(pbuf.String(), pbuf.Value())
		if err != nil {
			return err
		}
		if paren {
			i.WriteString(")")
		}
		return nil
	}

	if valuer, ok := value.(driver.Valuer); ok {
		// get driver.Valuer's data
		var err error
		value, err = valuer.Value()
		if err != nil {
			return err
		}
	}

	if value == nil {
		i.WriteString("NULL")
		return nil
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		i.WriteString(i.EncodeString(v.String()))
		return nil
	case reflect.Bool:
		i.WriteString(i.EncodeBool(v.Bool()))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i.WriteString(strconv.FormatInt(v.Int(), 10))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i.WriteString(strconv.FormatUint(v.Uint(), 10))
		return nil
	case reflect.Float32, reflect.Float64:
		i.WriteString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
		return nil
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			i.WriteString(i.EncodeTime(v.Interface().(time.Time)))
			return nil
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			i.WriteString(i.EncodeBytes(v.Bytes()))
			return nil
		}
		if v.Len() == 0 {
			// FIXME: support zero-length slice
			return ErrInvalidSliceLength
		}
		i.WriteString("(")
		for n := 0; n < v.Len(); n++ {
			if n > 0 {
				i.WriteString(",")
			}
			err := i.encodePlaceholder(v.Index(n).Interface())
			if err != nil {
				return err
			}
		}
		i.WriteString(")")
		return nil
	case reflect.Ptr:
		if v.IsNil() {
			i.WriteString("NULL")
			return nil
		}
		return i.encodePlaceholder(v.Elem().Interface())
	}
	return ErrNotSupported
}
