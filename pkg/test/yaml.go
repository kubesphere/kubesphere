package test

import (
	"bufio"
	"bytes"
	"io"

	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

const maxExecutiveEmpties = 100

// Scanner scans a yaml manifest file for manifest tokens delimited by "---".
// See bufio.Scanner for semantics.
type Scanner struct {
	reader  *k8syaml.YAMLReader
	token   []byte // Last token returned by split.
	err     error  // Sticky error.
	empties int    // Count of successive empty tokens.
	done    bool   // Scan has finished.
}

func NewYAMLScanner(b []byte) *Scanner {
	r := bufio.NewReader(bytes.NewBuffer(b))
	return &Scanner{reader: k8syaml.NewYAMLReader(r)}
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Scanner) Scan() bool {
	if s.done {
		return false
	}

	var (
		tok []byte
		err error
	)

	for {
		tok, err = s.reader.Read()
		if err != nil {
			if err == io.EOF {
				s.done = true
			}
			s.err = err
			return false
		}
		if len(bytes.TrimSpace(tok)) == 0 {
			s.empties++
			if s.empties > maxExecutiveEmpties {
				panic("yaml.Scan: too many empty tokens without progressing")
			}
			continue
		}
		s.empties = 0
		s.token = tok
		return true
	}
}

func (s *Scanner) Text() string {
	return string(s.token)
}

func (s *Scanner) Bytes() []byte {
	return s.token
}
