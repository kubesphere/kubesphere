package casdoor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlugin(t *testing.T) {
	var pub string = `-----BEGIN CERTIFICATE-----
MIIE+TCCAuGgAwIBAgIDAeJAMA0GCSqGSIb3DQEBCwUAMDYxHTAbBgNVBAoTFENh
c2Rvb3IgT3JnYW5pemF0aW9uMRUwEwYDVQQDEwxDYXNkb29yIENlcnQwHhcNMjEx
MDE1MDgxMTUyWhcNNDExMDE1MDgxMTUyWjA2MR0wGwYDVQQKExRDYXNkb29yIE9y
Z2FuaXphdGlvbjEVMBMGA1UEAxMMQ2FzZG9vciBDZXJ0MIICIjANBgkqhkiG9w0B
AQEFAAOCAg8AMIICCgKCAgEAsInpb5E1/ym0f1RfSDSSE8IR7y+lw+RJjI74e5ej
rq4b8zMYk7HeHCyZr/hmNEwEVXnhXu1P0mBeQ5ypp/QGo8vgEmjAETNmzkI1NjOQ
CjCYwUrasO/f/MnI1C0j13vx6mV1kHZjSrKsMhYY1vaxTEP3+VB8Hjg3MHFWrb07
uvFMCJe5W8+0rKErZCKTR8+9VB3janeBz//zQePFVh79bFZate/hLirPK0Go9P1g
OvwIoC1A3sarHTP4Qm/LQRt0rHqZFybdySpyWAQvhNaDFE7mTstRSBb/wUjNCUBD
PTSLVjC04WllSf6Nkfx0Z7KvmbPstSj+btvcqsvRAGtvdsB9h62Kptjs1Yn7GAuo
I3qt/4zoKbiURYxkQJXIvwCQsEftUuk5ew5zuPSlDRLoLByQTLbx0JqLAFNfW3g/
pzSDjgd/60d6HTmvbZni4SmjdyFhXCDb1Kn7N+xTojnfaNkwep2REV+RMc0fx4Gu
hRsnLsmkmUDeyIZ9aBL9oj11YEQfM2JZEq+RVtUx+wB4y8K/tD1bcY+IfnG5rBpw
IDpS262boq4SRSvb3Z7bB0w4ZxvOfJ/1VLoRftjPbLIf0bhfr/AeZMHpIKOXvfz4
yE+hqzi68wdF0VR9xYc/RbSAf7323OsjYnjjEgInUtRohnRgCpjIk/Mt2Kt84Kb0
wn8CAwEAAaMQMA4wDAYDVR0TAQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAgEAn2lf
DKkLX+F1vKRO/5gJ+Plr8P5NKuQkmwH97b8CS2gS1phDyNgIc4/LSdzuf4Awe6ve
C06lVdWSIis8UPUPdjmT2uMPSNjwLxG3QsrimMURNwFlLTfRem/heJe0Zgur9J1M
8haawdSdJjH2RgmFoDeE2r8NVRfhbR8KnCO1ddTJKuS1N0/irHz21W4jt4rxzCvl
2nR42Fybap3O/g2JXMhNNROwZmNjgpsF7XVENCSuFO1jTywLaqjuXCg54IL7XVLG
omKNNNcc8h1FCeKj/nnbGMhodnFWKDTsJcbNmcOPNHo6ixzqMy/Hqc+mWYv7maAG
Jtevs3qgMZ8F9Qzr3HpUc6R3ZYYWDY/xxPisuKftOPZgtH979XC4mdf0WPnOBLqL
2DJ1zaBmjiGJolvb7XNVKcUfDXYw85ZTZQ5b9clI4e+6bmyWqQItlwt+Ati/uFEV
XzCj70B4lALX6xau1kLEpV9O1GERizYRz5P9NJNA7KoO5AVMp9w0DQTkt+LbXnZE
HHnWKy8xHQKZF9sR7YBPGLs/Ac6tviv5Ua15OgJ/8dLRZ/veyFfGo2yZsI+hKVU5
nCCJHBcAyFnm1hdvdwEdH33jDBjNB6ciotJZrf/3VYaIWSalADosHAgMWfXuWP+h
8XKXmzlxuHbTMQYtZPDgspS5aK+S4Q9wb8RRAYo=
-----END CERTIFICATE-----`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		switch r.RequestURI {
		case "/api/login/oauth/access_token":
			data = map[string]interface{}{
				"access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJvd25lciI6ImJ1aWx0LWluIiwibmFtZSI6ImFkbWluIiwiY3JlYXRlZFRpbWUiOiIyMDIxLTEwLTI5VDA4OjE4OjIwWiIsInVwZGF0ZWRUaW1lIjoiIiwiaWQiOiI0YWM5NzNjMC00YzRiLTQ0ZTMtOGU2ZS04NGY0YTE3YjEzMGYiLCJ0eXBlIjoibm9ybWFsLXVzZXIiLCJwYXNzd29yZCI6IiIsInBhc3N3b3JkU2FsdCI6IiIsImRpc3BsYXlOYW1lIjoiQWRtaW4iLCJhdmF0YXIiOiJodHRwczovL2Nhc2Jpbi5vcmcvaW1nL2Nhc2Jpbi5zdmciLCJwZXJtYW5lbnRBdmF0YXIiOiIiLCJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwicGhvbmUiOiIxMjM0NTY3ODkxMCIsImxvY2F0aW9uIjoiIiwiYWRkcmVzcyI6W10sImFmZmlsaWF0aW9uIjoiRXhhbXBsZSBJbmMuIiwidGl0bGUiOiIiLCJpZENhcmRUeXBlIjoiIiwiaWRDYXJkIjoiIiwiaG9tZXBhZ2UiOiIiLCJiaW8iOiIiLCJ0YWciOiJzdGFmZiIsInJlZ2lvbiI6IiIsImxhbmd1YWdlIjoiIiwiZ2VuZGVyIjoiIiwiYmlydGhkYXkiOiIiLCJlZHVjYXRpb24iOiIiLCJzY29yZSI6MCwicmFua2luZyI6MCwiaXNPbmxpbmUiOmZhbHNlLCJpc0FkbWluIjp0cnVlLCJpc0dsb2JhbEFkbWluIjp0cnVlLCJpc0ZvcmJpZGRlbiI6ZmFsc2UsImlzRGVsZXRlZCI6ZmFsc2UsInNpZ251cEFwcGxpY2F0aW9uIjoiIiwiaGFzaCI6IjE1YTFjMTJiYWIxNDk1MDY2NGNmMmEyM2NkNjJhN2E3IiwicHJlSGFzaCI6IjE1YTFjMTJiYWIxNDk1MDY2NGNmMmEyM2NkNjJhN2E3IiwiY3JlYXRlZElwIjoiIiwibGFzdFNpZ25pblRpbWUiOiIiLCJsYXN0U2lnbmluSXAiOiIiLCJnaXRodWIiOiIiLCJnb29nbGUiOiIiLCJxcSI6IiIsIndlY2hhdCI6IiIsImZhY2Vib29rIjoiIiwiZGluZ3RhbGsiOiIiLCJ3ZWlibyI6IiIsImdpdGVlIjoiIiwibGlua2VkaW4iOiIiLCJ3ZWNvbSI6IiIsImxhcmsiOiIiLCJnaXRsYWIiOiIiLCJsZGFwIjoiIiwicHJvcGVydGllcyI6e30sImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODAwMCIsInN1YiI6IjRhYzk3M2MwLTRjNGItNDRlMy04ZTZlLTg0ZjRhMTdiMTMwZiIsImF1ZCI6WyI4OGIyNDU3YTEyMzk4NGI0ODM5MiJdLCJleHAiOjE2MzgxNzM0NjYsIm5iZiI6MTYzNzU2ODY2NiwiaWF0IjoxNjM3NTY4NjY2fQ.ZwKN4Q82yjfuEKUh0OB0sO6K1CIpBQ8_SgGbBVFpddmqtzxoRsmKI4x2fyfFrApflsyGEthuj8ZUR77Fvg59_hQIFk8F6E0ToBRYB96AbXQP9YWdvaxRbCjcii2cNNqU4dw8vpf_k3eZaDzoKT8bIeHJd1Bt4x9xgBPxQi_8PydLxrSTyhlFeDScrbmUyOjV-bFx5ERmRMgml7q4RraN2-6w0aFKfE1BJiGZurnwcTBdUx1fctbgiIeO77QCOSJk0RiXj5M72nvjWfcUlTWrEvRXgtToaUpGkghnfUCJbg60_8Dvqb-y_M_xl7Wj2ydVJ8eeq7aw5mlSi_H91UJ4O6f8eYlLygvozdsWIv5K801tCb5kFacwsf494XA9rlkCOmF8CkGJK3G_054LFN1Hmt9kav-9MheaOjdLAH-zvgNqrjoWulIR2qtNdMM_zh6ZZzJoq8RwGfYCajzFeVYbtG4x5b_y7MctIE7JgpC9X_Y4Nu7kLL_jplFfwfoDSqYllNsq0u-EF6I_pY1CO2EqlxY8jg1t09Es8FaCU4CW3CjC_CniGx6V348dLmZCFTc9SDPGuLrnJGlawCIlczf0FrQm11o3O32n4M_LVv-LdjScIGUjlagTfAY3jNNDRCjXDPAt4i8yBrOciJykF-WU7IkQxwivqd5fzHhSq1CP1rw",
				"expires_in":   10086,
				"scope":        "read",
				"token_type":   "bearer",
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("not implemented"))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	var code = "81d464c9f166e758c41d"
	var state = "5577006791947779410"

	request, err := http.NewRequest("GET", fmt.Sprintf("http://example.com?code=%s&state=%s", code, state), nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	config := oauth.DynamicOptions{
		"jwtSecret":        pub,
		"casdoorEndPoint":  server.URL,
		"organizationName": "xxx",
		"clientID":         "xxx",
		"clientSecret":     "xxx",
		"applicationName":  "xxx",
	}
	factory := CasdoorOAuthProviderFactory{}
	provider, err := factory.Create(config)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	identity, err := provider.IdentityExchangeCallback(request)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if identity.GetUsername() != "admin" {
		t.Errorf("identity.GetUsername() is %s", identity.GetUsername())
		return
	}
	if identity.GetUserID() != "4ac973c0-4c4b-44e3-8e6e-84f4a17b130f" {
		t.Errorf("identity.GetUserID() is %s", identity.GetUserID())
		return
	}
	if identity.GetEmail() != "admin@example.com" {
		t.Errorf("identity.GetEmail() is %s", identity.GetEmail())
		return
	}

}
