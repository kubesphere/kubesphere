/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package iputil

import (
	"net/http"
	"testing"
)

func TestRemoteIp(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    http.Header
		want       string
	}{
		{
			name:       "uses remote address host",
			remoteAddr: "192.0.2.10:12345",
			want:       "192.0.2.10",
		},
		{
			name:       "normalizes IPv6 localhost",
			remoteAddr: "[::1]:12345",
			want:       "127.0.0.1",
		},
		{
			name:       "x-client-ip has highest priority",
			remoteAddr: "192.0.2.10:12345",
			headers: http.Header{
				"X-Client-Ip":     []string{"203.0.113.10"},
				"X-Real-Ip":       []string{"203.0.113.20"},
				"X-Forwarded-For": []string{"203.0.113.30"},
			},
			want: "203.0.113.10",
		},
		{
			name:       "x-forwarded-for returns original client",
			remoteAddr: "192.0.2.10:12345",
			headers: http.Header{
				"X-Forwarded-For": []string{"203.0.113.10, 10.0.0.2"},
			},
			want: "203.0.113.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     tt.headers,
			}

			if got := RemoteIp(req); got != tt.want {
				t.Fatalf("RemoteIp() = %q, want %q", got, tt.want)
			}
		})
	}
}
