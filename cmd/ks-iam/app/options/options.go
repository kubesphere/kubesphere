/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package options

import (
	"github.com/spf13/pflag"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	AdminEmail              string
	AdminPassword           string
	TokenExpireTime         string
	JWTSecret               string
	AuthRateLimit           string
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
	}
	return s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.AdminEmail, "admin-email", "admin@kubesphere.io", "default administrator's email")
	fs.StringVar(&s.AdminPassword, "admin-password", "passw0rd", "default administrator's password")
	fs.StringVar(&s.TokenExpireTime, "token-expire-time", "2h", "token expire time,valid time units are \"ns\",\"us\",\"ms\",\"s\",\"m\",\"h\"")
	fs.StringVar(&s.JWTSecret, "jwt-secret", "", "jwt secret")
	fs.StringVar(&s.AuthRateLimit, "auth-rate-limit", "5/30m", "specifies the maximum number of authentication attempts permitted and time interval,valid time units are \"s\",\"m\",\"h\"")
	s.GenericServerRunOptions.AddFlags(fs)
}
