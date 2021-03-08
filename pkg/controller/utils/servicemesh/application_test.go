package servicemesh

import "testing"

func TestIsHTTP(t *testing.T) {
	if !SupportHttpProtocol("gRPC") {
		t.Errorf("gRPC is HTTP protocol")
	}
	if !SupportHttpProtocol("HTTP") {
		t.Errorf("HTTP is HTTP protocol")
	}
	if !SupportHttpProtocol("HTTP2") {
		t.Errorf("HTTP2 is HTTP protocol")
	}
	if SupportHttpProtocol("Mysql") {
		t.Errorf("mysql is not HTTP protocol")
	}
	if SupportHttpProtocol("udp") {
		t.Errorf("UDP is not HTTP protocol")
	}
}
