package handshake

// This package uses unsafe to convert between:
// * qtls.ConnectionState and tls.ConnectionState
// * qtls.ClientSessionState and tls.ClientSessionState
// We check in init() that this conversion actually is safe.

import (
	"crypto/tls"
	"reflect"

	"github.com/marten-seemann/qtls"
)

func init() {
	if !structsEqual(&tls.ConnectionState{}, &qtls.ConnectionState{}) {
		panic("qtls.ConnectionState not compatible with tls.ConnectionState")
	}
	if !structsEqual(&tls.ClientSessionState{}, &qtls.ClientSessionState{}) {
		panic("qtls.ClientSessionState not compatible with tls.ClientSessionState")
	}
}

func structsEqual(a, b interface{}) bool {
	sa := reflect.ValueOf(a).Elem()
	sb := reflect.ValueOf(b).Elem()
	if sa.NumField() != sb.NumField() {
		return false
	}
	for i := 0; i < sa.NumField(); i++ {
		fa := sa.Type().Field(i)
		fb := sb.Type().Field(i)
		if !reflect.DeepEqual(fa.Index, fb.Index) || fa.Name != fb.Name || fa.Anonymous != fb.Anonymous || fa.Offset != fb.Offset || !reflect.DeepEqual(fa.Type, fb.Type) {
			return false
		}
	}
	return true
}
