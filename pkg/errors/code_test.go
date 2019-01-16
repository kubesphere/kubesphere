package errors

import "testing"

func TestCode_String(t *testing.T) {
	t.Log(Code(1).String())
}
