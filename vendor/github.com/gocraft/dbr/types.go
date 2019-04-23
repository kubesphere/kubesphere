package dbr

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

//
// Your app can use these Null types instead of the defaults. The sole benefit you get is a MarshalJSON method that is not retarded.
//

// NullString is a type that can be null or a string
type NullString struct {
	sql.NullString
}

// NullFloat64 is a type that can be null or a float64
type NullFloat64 struct {
	sql.NullFloat64
}

// NullInt64 is a type that can be null or an int
type NullInt64 struct {
	sql.NullInt64
}

// NullTime is a type that can be null or a time
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

// NullBool is a type that can be null or a bool
type NullBool struct {
	sql.NullBool
}

var nullString = []byte("null")

// MarshalJSON correctly serializes a NullString to JSON
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullInt64 to JSON
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int64)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullFloat64 to JSON
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Float64)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullTime to JSON
func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullBool to JSON
func (n NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Bool)
	}
	return nullString, nil
}

// UnmarshalJSON correctly deserializes a NullString from JSON
func (n *NullString) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullInt64 from JSON
func (n *NullInt64) UnmarshalJSON(b []byte) error {
	var s json.Number
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		return n.Scan(nil)
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullFloat64 from JSON
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullTime from JSON
func (n *NullTime) UnmarshalJSON(b []byte) error {
	// scan for null
	if bytes.Equal(b, nullString) {
		return n.Scan(nil)
	}
	// scan for JSON timestamp
	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	return n.Scan(t)
}

// UnmarshalJSON correctly deserializes a NullBool from JSON
func (n *NullBool) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

func NewNullInt64(v interface{}) (n NullInt64) {
	n.Scan(v)
	return
}

func NewNullFloat64(v interface{}) (n NullFloat64) {
	n.Scan(v)
	return
}

func NewNullString(v interface{}) (n NullString) {
	n.Scan(v)
	return
}

func NewNullTime(v interface{}) (n NullTime) {
	n.Scan(v)
	return
}

func NewNullBool(v interface{}) (n NullBool) {
	n.Scan(v)
	return
}

// The `(*NullTime) Scan(interface{})` and `parseDateTime(string, *time.Location)`
// functions are slightly modified versions of code from the github.com/go-sql-driver/mysql
// package. They work with Postgres and MySQL databases. Potential future
// drivers should ensure these will work for them, or come up with an alternative.
//
// Conforming with its licensing terms the copyright notice and link to the licence
// are available below.
//
// Source: https://github.com/go-sql-driver/mysql/blob/527bcd55aab2e53314f1a150922560174b493034/utils.go#L452-L508

// Copyright notice from original developers:
//
// Go MySQL Driver - A MySQL-Driver for Go's database/sql package
//
// Copyright 2012 The Go-MySQL-Driver Authors. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/

// Scan implements the Scanner interface.
// The value type must be time.Time or string / []byte (formatted time-string),
// otherwise Scan fails.
func (n *NullTime) Scan(value interface{}) error {
	var err error

	if value == nil {
		n.Time, n.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		n.Time, n.Valid = v, true
		return nil
	case []byte:
		n.Time, err = parseDateTime(string(v), time.UTC)
		n.Valid = (err == nil)
		return err
	case string:
		n.Time, err = parseDateTime(v, time.UTC)
		n.Valid = (err == nil)
		return err
	}

	n.Valid = false
	return nil
}

func parseDateTime(str string, loc *time.Location) (time.Time, error) {
	var t time.Time
	var err error

	base := "0000-00-00 00:00:00.0000000"
	switch len(str) {
	case 10, 19, 21, 22, 23, 24, 25, 26:
		if str == base[:len(str)] {
			return t, err
		}
		t, err = time.Parse(timeFormat[:len(str)], str)
	default:
		err = ErrInvalidTimestring
		return t, err
	}

	// Adjust location
	if err == nil && loc != time.UTC {
		y, mo, d := t.Date()
		h, mi, s := t.Clock()
		t, err = time.Date(y, mo, d, h, mi, s, t.Nanosecond(), loc), nil
	}

	return t, err
}
