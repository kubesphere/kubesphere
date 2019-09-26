// Copyright 2017 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package db

import (
	"strings"

	"github.com/gocraft/dbr"
)

const (
	placeholder = "?"
)

type EqCondition struct {
	dbr.Builder
	Column string
	Value  interface{}
}

// Copy From vendor/github.com/gocraft/dbr/condition.go:36
func buildCmp(d dbr.Dialect, buf dbr.Buffer, pred string, column string, value interface{}) error {
	buf.WriteString(d.QuoteIdent(column))
	buf.WriteString(" ")
	buf.WriteString(pred)
	buf.WriteString(" ")
	buf.WriteString(placeholder)

	buf.WriteValue(value)
	return nil
}

// And creates AND from a list of conditions
func And(cond ...dbr.Builder) dbr.Builder {
	return dbr.And(cond...)
}

// Or creates OR from a list of conditions
func Or(cond ...dbr.Builder) dbr.Builder {
	return dbr.Or(cond...)
}

func escape(str string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '%', '\'', '^', '[', ']', '!', '_':
			return ' '
		}
		return r
	}, str)
}

func Like(column string, value string) dbr.Builder {
	value = "%" + strings.TrimSpace(escape(value)) + "%"
	return dbr.BuildFunc(func(d dbr.Dialect, buf dbr.Buffer) error {
		return buildCmp(d, buf, "LIKE", column, value)
	})
}

func Prefix(column string, value string) dbr.Builder {
	value = strings.TrimSpace(escape(value)) + "%"
	return dbr.BuildFunc(func(d dbr.Dialect, buf dbr.Buffer) error {
		return buildCmp(d, buf, "LIKE", column, value)
	})
}

// Eq is `=`.
// When value is nil, it will be translated to `IS NULL`.
// When value is a slice, it will be translated to `IN`.
// Otherwise it will be translated to `=`.
func Eq(column string, value interface{}) dbr.Builder {
	return &EqCondition{
		Builder: dbr.Eq(column, value),
		Column:  column,
		Value:   value,
	}
}

// Neq is `!=`.
// When value is nil, it will be translated to `IS NOT NULL`.
// When value is a slice, it will be translated to `NOT IN`.
// Otherwise it will be translated to `!=`.
func Neq(column string, value interface{}) dbr.Builder {
	return dbr.Neq(column, value)
}

// Gt is `>`.
func Gt(column string, value interface{}) dbr.Builder {
	return dbr.Gt(column, value)
}

// Gte is '>='.
func Gte(column string, value interface{}) dbr.Builder {
	return dbr.Gte(column, value)
}

// Lt is '<'.
func Lt(column string, value interface{}) dbr.Builder {
	return dbr.Lt(column, value)
}

// Lte is `<=`.
func Lte(column string, value interface{}) dbr.Builder {
	return dbr.Lte(column, value)
}
