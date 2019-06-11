package dbr

import (
	"database/sql"
	"reflect"
)

// Load loads any value from sql.Rows
func Load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()

	column, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrInvalidPointer
	}
	v = v.Elem()
	isScanner := v.Addr().Type().Implements(typeScanner)
	isSlice := v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 && !isScanner
	isMap := v.Kind() == reflect.Map && !isScanner
	isMapOfSlices := isMap && v.Type().Elem().Kind() == reflect.Slice && v.Type().Elem().Elem().Kind() != reflect.Uint8
	if isMap {
		v.Set(reflect.MakeMap(v.Type()))
	}
	count := 0
	for rows.Next() {
		var elem, keyElem reflect.Value
		var ptr []interface{}
		var err error

		if isMapOfSlices {
			elem = reflect.New(v.Type().Elem().Elem()).Elem()
		} else if isSlice || isMap {
			elem = reflect.New(v.Type().Elem()).Elem()
		} else {
			elem = v
		}

		if isMap {
			ptr, err = findPtr(column[1:], elem)
			if err != nil {
				return 0, err
			}
			keyElem = reflect.New(v.Type().Key()).Elem()
			keyPtr, err := findPtr(column[0:1], keyElem)
			if err != nil {
				return 0, err
			}
			ptr = append(keyPtr, ptr...)
		} else {
			ptr, err = findPtr(column, elem)
			if err != nil {
				return 0, err
			}
		}

		err = rows.Scan(ptr...)
		if err != nil {
			return 0, err
		}

		count++

		if isSlice {
			v.Set(reflect.Append(v, elem))
		} else if isMapOfSlices {
			s := v.MapIndex(keyElem)
			if !s.IsValid() {
				s = reflect.Zero(v.Type().Elem())
			}
			v.SetMapIndex(keyElem, reflect.Append(s, elem))
		} else if isMap {
			v.SetMapIndex(keyElem, elem)
		} else {
			break
		}
	}
	return count, nil
}

type dummyScanner struct{}

func (dummyScanner) Scan(interface{}) error {
	return nil
}

var (
	dummyDest   sql.Scanner = dummyScanner{}
	typeScanner             = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func findPtr(column []string, value reflect.Value) ([]interface{}, error) {
	if value.Addr().Type().Implements(typeScanner) {
		return []interface{}{value.Addr().Interface()}, nil
	}
	switch value.Kind() {
	case reflect.Struct:
		var ptr []interface{}
		m := structMap(value)
		for _, key := range column {
			if val, ok := m[key]; ok {
				ptr = append(ptr, val.Addr().Interface())
			} else {
				ptr = append(ptr, dummyDest)
			}
		}
		return ptr, nil
	case reflect.Ptr:
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return findPtr(column, value.Elem())
	}
	return []interface{}{value.Addr().Interface()}, nil
}
