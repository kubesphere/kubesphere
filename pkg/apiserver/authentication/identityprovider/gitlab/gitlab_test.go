package gitlab

import (
	"reflect"
	"testing"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

func Test_gitlabProviderFactory_Create(t *testing.T) {
	type args struct {
		opts options.DynamicOptions
	}

	mustUnmarshalYAML := func(data string) options.DynamicOptions {
		var dynamicOptions options.DynamicOptions
		_ = yaml.Unmarshal([]byte(data), &dynamicOptions)
		return dynamicOptions
	}

	tests := []struct {
		name    string
		args    args
		want    identityprovider.OAuthProvider
		wantErr bool
	}{
		{
			name: "should create successfully",
			args: args{opts: mustUnmarshalYAML(`
clientID: 035c18fc229c686e4652d7034
clientSecret: 75c82b42e54aaf25186140f5
endpoint:
  userInfoUrl: "https://gitlab.com/api/v4/user"
  authURL: "https://gitlab.com/oauth/authorize"
  tokenURL: "https://gitlab.com/oauth/token"
redirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/gitlab"
scopes:
- read
`)},
			want: &gitlab{
				ClientID:     "035c18fc229c686e4652d7034",
				ClientSecret: "75c82b42e54aaf25186140f5",
				Endpoint: endpoint{
					AuthURL:     "https://gitlab.com/oauth/authorize",
					TokenURL:    "https://gitlab.com/oauth/token",
					UserInfoURL: "https://gitlab.com/api/v4/user",
				},
				RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/gitlab",
				Scopes:      []string{"read"},
				Config: &oauth2.Config{
					ClientID:     "035c18fc229c686e4652d7034",
					ClientSecret: "75c82b42e54aaf25186140f5",
					Endpoint: oauth2.Endpoint{
						AuthURL:   "https://gitlab.com/oauth/authorize",
						TokenURL:  "https://gitlab.com/oauth/token",
						AuthStyle: oauth2.AuthStyleAutoDetect,
					},
					RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/gitlab",
					Scopes:      []string{"read"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gitlabProviderFactory{}
			got, err := g.Create(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}
