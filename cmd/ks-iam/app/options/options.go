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
	"flag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	genericoptions "kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"strings"
	"time"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	KubernetesOptions       *k8s.KubernetesOptions
	LdapOptions             *ldap.Options
	RedisOptions            *cache.Options
	MySQLOptions            *mysql.Options
	AdminEmail              string
	AdminPassword           string
	TokenIdleTimeout        time.Duration
	JWTSecret               string
	AuthRateLimit           string
	EnableMultiLogin        bool
	GenerateKubeConfig      bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		KubernetesOptions:       k8s.NewKubernetesOptions(),
		LdapOptions:             ldap.NewLdapOptions(),
		MySQLOptions:            mysql.NewMySQLOptions(),
		RedisOptions:            cache.NewRedisOptions(),
	}
	return s
}

func (s *ServerRunOptions) Flags(conf *ServerRunOptions) (fss cliflag.NamedFlagSets) {

	fs := fss.FlagSet("generic")

	s.GenericServerRunOptions.AddFlags(fs, conf.GenericServerRunOptions)
	fs.StringVar(&s.AdminEmail, "admin-email", "admin@kubesphere.io", "default administrator's email")
	fs.StringVar(&s.AdminPassword, "admin-password", "passw0rd", "default administrator's password")
	fs.DurationVar(&s.TokenIdleTimeout, "token-idle-timeout", 30*time.Minute, "tokens that are idle beyond that time will expire,0s means the token has no expiration time. valid time units are \"ns\",\"us\",\"ms\",\"s\",\"m\",\"h\"")
	fs.StringVar(&s.JWTSecret, "jwt-secret", "", "jwt secret")
	fs.StringVar(&s.AuthRateLimit, "auth-rate-limit", "5/30m", "specifies the maximum number of authentication attempts permitted and time interval,valid time units are \"s\",\"m\",\"h\"")
	fs.BoolVar(&s.EnableMultiLogin, "enable-multi-login", false, "allow one account to have multiple sessions")
	fs.BoolVar(&s.GenerateKubeConfig, "generate-kubeconfig", true, "generate kubeconfig for new users, kubeconfig is required in devops pipeline, set to false if you don't need devops.")

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), conf.KubernetesOptions)
	s.LdapOptions.AddFlags(fss.FlagSet("ldap"), conf.LdapOptions)
	s.RedisOptions.AddFlags(fss.FlagSet("redis"), conf.RedisOptions)
	s.MySQLOptions.AddFlags(fss.FlagSet("mysql"), conf.MySQLOptions)

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}
