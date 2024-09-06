/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package token

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
	"k8s.io/apiserver/pkg/authentication/user"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

const privateKeyData = `
-----BEGIN RSA PRIVATE KEY-----
MIIEoQIBAAKCAQEAnDK2bNmX+tBWY/JHll1T1LF3/6RTbJ2qUsvwZVuVP/XbmWeY
vDZyTR+YL6JaqRC/NibphgCV0p6cKZNuoGCEHpS0Ix9ZZwkA8BhwrFwAU0O1Qmrv
v7It3p0Lc9WKN7PBWDQnUIdeSWnAeSbmWETP8Y2e+vG/iusLojPJenEaiOiwzU8p
3CGSh7IPBXF3aUeB3dgJuaiumDuzlp0Oe/xKvWo0faB2hFXi36KaLMpugNcbejKl
R6w3jH5wJjto5XqTEpW4a77K4rt7CFXVGfcbLo+n/5j3oC0lw4KOy7OX0Qf1jY+x
sa1Q+3UoDC02sQRf77uj3eITol8Spoo7wfJqmwIDAQABAoIBAGEArYIT7+p3j+8p
+4NKGlGwlRFR/+0oTSp2NKj9o0bBbMtsJtJcDcgPoveSIDN2jwkWSVhK7MCMd/bp
9H3s8p/7QZO+WEtAsDBrPS4NRLZxChRhTNsD0LC7Xu1k5B2LqLsaSIAeUVPONRYI
Lm0K7wjYJq85iva+2c610p4Tt6LlxuOu41Zw7RAaW8nBoMdQzi19X+hUloogVo7S
hid8gm2KUPY6xF+RpHGQ5OUND0d+2wBkHxbYNRIfYrxCKt8+dLykLzAmm+ScCfyG
jKcNoRwW5s/3ttR7r7hn3whttydkper5YvxM3+EvL83H7JL11KHcHy/yPYv+2IxQ
psvEtIECgYEAykCm/w58pdifLuWG1mwHCkdQ6wCd9gAHIDfaqbFhTcdwGGYXb1xQ
3CHjkkX6rpB3rJReCxlGq01OemVNlpIGLhdnK87aX5pRVn2fHGaMoC2V5RWv3pyE
3gJ41h9FtPX2juKFG9PNiR7FrtKPzQczfh2L1OMpLOXfPgxvo/fXBQsCgYEAxbTz
mibb4F/TBVXMuSL7Pk9hBPlFgFIEUKbiqt+sKQCqSZPGjV5giDfQDGsJ5EqOkRg0
qlCrKk+DW+d+Pmc4yo58cd2xPnQETholV19+dM3AGiy4BMjeUnJD+Dme7M/fhrlW
IK/1ZErKSZ3nN20qeneIFltm6+4pgQ1HB9KwirECgYAy65wf0xHm32cUc41DJueO
2u2wfPNIIDGrFuTinFoXLwM14V49F0z0X0Pga+X1VUIMHT6gJLj6H/iGMEMciZ8s
s4+yI94u+7dGw1Hv4JG/Mjru9krVDSsWiiDKKA1wxgxRZQ6GNwkkYK78mN7Di/CW
6/Fso9SWDTnrcU4aRifIiQKBgQCQ+kJwVfKCtIIPtX0sfeRzKs5gUVKP6JTVd6tb
1i1u29gDoGPHIt/yw8rCcHOOfsXQzElCY2lA25HeAQFoTVUt5BKJhSIGRBksFKwx
SAt5J6+pAgXnLE0rdDM3gTlzOnQVXS81RRLTeqygEzSMRncR2zll+5ybgcfZpJzj
tbJT4QJ/Y02wfkm1dL/BFg520/otVeuC+Bt+YyWMVs867xLLzFci7tj6ZzlzMorQ
PsSsOHhPx0g+Wl8K2+Edg3FQRZ1m0rQFAZn66jd96u85aA9NH/bw3A3VYUdVJyHh
4ZgZLx9JMCkmRfa7Dp2mzoqGUC1cjNvm722baeMqXpHSXDP2Jg==
-----END RSA PRIVATE KEY-----
`

func TestNewIssuer(t *testing.T) {
	signKeyData := base64.StdEncoding.EncodeToString([]byte(privateKeyData))
	config := &oauth.IssuerOptions{
		URL:              "https://ks-console.kubesphere-system.svc",
		SignKeyData:      signKeyData,
		MaximumClockSkew: 10 * time.Second,
		JWTSecret:        "test-secret",
	}
	got, err := NewIssuer(config)
	if err != nil {
		t.Fatal(err)
	}

	signKey, keyID, err := loadSignKey(config)
	if err != nil {
		t.Fatal(err)
	}

	want := &issuer{
		url:              config.URL,
		secret:           []byte(config.JWTSecret),
		maximumClockSkew: config.MaximumClockSkew,
		signKey: &Keys{
			SigningKey: &jose.JSONWebKey{
				Key:       signKey,
				KeyID:     keyID,
				Algorithm: jwt.SigningMethodRS256.Alg(),
				Use:       "sig",
			},
			SigningKeyPub: &jose.JSONWebKey{
				Key:       signKey.Public(),
				KeyID:     keyID,
				Algorithm: jwt.SigningMethodRS256.Alg(),
				Use:       "sig",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewIssuerOptions() got = %v, want %v", got, want)
		return
	}
}

func TestNewIssuerGenerateSignKey(t *testing.T) {
	config := &oauth.IssuerOptions{
		URL:              "https://ks-console.kubesphere-system.svc",
		MaximumClockSkew: 10 * time.Second,
		JWTSecret:        "test-secret",
	}

	got, err := NewIssuer(config)
	if err != nil {
		t.Fatal(err)
	}

	iss := got.(*issuer)
	assert.NotNil(t, iss.signKey)
	assert.NotNil(t, iss.signKey.SigningKey)
	assert.NotNil(t, iss.signKey.SigningKeyPub)
	assert.NotNil(t, iss.signKey.SigningKey.KeyID)
	assert.NotNil(t, iss.signKey.SigningKeyPub.KeyID)
}

func Test_issuer_IssueTo(t *testing.T) {
	type fields struct {
		url              string
		secret           []byte
		maximumClockSkew time.Duration
	}
	type args struct {
		request *IssueRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VerifiedResponse
		wantErr bool
	}{
		{
			name: "token is successfully issued",
			fields: fields{
				url:              "kubesphere",
				secret:           []byte("kubesphere"),
				maximumClockSkew: 0,
			},
			args: args{request: &IssueRequest{
				User: &user.DefaultInfo{
					Name: "user1",
				},
				Claims: Claims{
					TokenType: AccessToken,
				},
				ExpiresIn: 2 * time.Hour},
			},
			want: &VerifiedResponse{
				User: &user.DefaultInfo{
					Name: "user1",
				},
				Claims: Claims{
					Username:  "user1",
					TokenType: AccessToken,
				},
			},
			wantErr: false,
		},
		{
			name: "token is successfully issued",
			fields: fields{
				url:              "kubesphere",
				secret:           []byte("kubesphere"),
				maximumClockSkew: 0,
			},
			args: args{request: &IssueRequest{
				User: &user.DefaultInfo{
					Name: "user2",
				},
				Claims: Claims{
					Username:  "user2",
					TokenType: RefreshToken,
				},
				ExpiresIn: 0},
			},
			want: &VerifiedResponse{
				User: &user.DefaultInfo{
					Name: "user2",
				},
				Claims: Claims{
					Username:  "user2",
					TokenType: RefreshToken,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &issuer{
				url:              tt.fields.url,
				secret:           tt.fields.secret,
				maximumClockSkew: tt.fields.maximumClockSkew,
			}
			token, err := s.IssueTo(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("IssueTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := s.Verify(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}
			assert.Equal(t, got.TokenType, tt.want.TokenType)
			assert.Equal(t, got.Issuer, tt.fields.url)
			assert.Equal(t, got.Username, tt.want.Username)
			assert.Equal(t, got.Subject, tt.want.User.GetName())
			assert.NotZero(t, got.IssuedAt)
		})
	}
}

func Test_issuer_Verify(t *testing.T) {
	type fields struct {
		url              string
		secret           []byte
		maximumClockSkew time.Duration
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VerifiedResponse
		wantErr bool
	}{
		{
			name: "token validation failed",
			fields: fields{
				url:              "kubesphere",
				secret:           []byte("kubesphere"),
				maximumClockSkew: 0,
			},
			args:    args{token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwidG9rZW5fdHlwZSI6ImFjY2Vzc190b2tlbiIsImV4cCI6MTYzMDY0MDMyMywiaWF0IjoxNjMwNjM2NzIzLCJpc3MiOiJrdWJlc3BoZXJlIiwibmJmIjoxNjMwNjM2NzIzfQ.4ENxyPTIe-BoQfuY5F4Mon5tB3KeV06B4i2JITRlPA8"},
			wantErr: true,
		},
		{
			name: "token is successfully verified",
			fields: fields{
				url:              "kubesphere",
				secret:           []byte("kubesphere"),
				maximumClockSkew: 0,
			},
			args: args{token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2MzA2MzczOTgsImlzcyI6Imt1YmVzcGhlcmUiLCJzdWIiOiJ1c2VyMiIsInRva2VuX3R5cGUiOiJyZWZyZXNoX3Rva2VuIiwidXNlcm5hbWUiOiJ1c2VyMiJ9.vqPczw4SyytVOQmgaK9ip2dvg2fSQStUUE_Y7Ts45WY"},
			want: &VerifiedResponse{
				User: &user.DefaultInfo{
					Name: "user2",
				},
				Claims: Claims{
					Username:  "user2",
					TokenType: RefreshToken,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &issuer{
				url:              tt.fields.url,
				secret:           tt.fields.secret,
				maximumClockSkew: tt.fields.maximumClockSkew,
			}
			got, err := s.Verify(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}
			assert.Equal(t, got.TokenType, tt.want.TokenType)
			assert.Equal(t, got.Issuer, tt.fields.url)
			assert.Equal(t, got.Username, tt.want.Username)
			assert.Equal(t, got.Subject, tt.want.User.GetName())
			assert.NotZero(t, got.IssuedAt)
		})
	}
}

func Test_issuer_keyFunc(t *testing.T) {
	type fields struct {
		//nolint:unused
		name   string
		secret []byte
		//nolint:unused
		maximumClockSkew time.Duration
	}
	type args struct {
		token *jwt.Token
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "sign key obtained successfully",
			fields: fields{
				secret: []byte("kubesphere"),
			},
			args: args{token: &jwt.Token{
				Method: jwt.SigningMethodHS256,
				Header: map[string]interface{}{"alg": "HS256"},
			}},
		},
		{
			name:   "sign key obtained successfully",
			fields: fields{},
			args: args{token: &jwt.Token{
				Method: jwt.SigningMethodRS256,
				Header: map[string]interface{}{"alg": "RS256"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewIssuer(oauth.NewIssuerOptions())
			if err != nil {
				t.Error(err)
				return
			}
			iss := s.(*issuer)
			got, _ := iss.keyFunc(tt.args.token)
			assert.NotNil(t, got)
		})
	}
}
