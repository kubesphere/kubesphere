// Package csvutil provides fast and idiomatic mapping between CSV and Go values.
//
// This package does not provide a CSV parser itself, it is based on the Reader and Writer
// interfaces which are implemented by eg. std csv package. This gives a possibility
// of choosing any other CSV writer or reader which may be more performant.
package csvutil
