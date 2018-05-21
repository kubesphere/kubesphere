/*
Copyright 2018 The KubeSphere Authors.

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

package models

import (
	"bytes"
	"github.com/golang/glog"
	"text/template"
	"encoding/base64"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/options"
)

var kubeconfigTemp =
	"apiVersion: v1\n" +
	"clusters:\n" +
	"- cluster:\n" +
	"    certificate-authority-data: {{.Certificate}}\n" +
	"    server: {{.Server}}\n" +
	"    name: kubernetes\n" +
	"contexts:\n" +
	"- context:\n" +
	"    cluster: kubernetes\n" +
	"    user: {{.User}}\n" +
	"    namespace: {{.User}}\n" +
	"  name: default\n" +
	"current-context: default\n" +
	"kind: Config\n" +
	"preferences: {}\n" +
	"users:\n" +
	"- name: {{.User}}\n" +
	"  user:\n" +
	"  token: {{.Token}}\n"

const DefaultServiceAccount = "default"

type Config struct {
	Certificate string
	Server    string
	User string
	Token string
}

func GetKubeConfig(namespace string) (string, error) {
	tmpl, err := template.New("").Parse(kubeconfigTemp)
	if err != nil {
		glog.Errorln(err)
		return "", err
	}

	kubeConfig, err := getKubeConfig(namespace, options.ServerOptions.GetApiServerHost())

	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, kubeConfig)
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	return buf.String(), nil
}

func getKubeConfig(namespace, apiserverHost string) (*Config, error) {
	k8sClient := client.NewK8sClient()
	saInfo, err := k8sClient.CoreV1().ServiceAccounts(namespace).Get(DefaultServiceAccount, meta_v1.GetOptions{})
	if err != nil{
		glog.Errorln(err)
		return nil, err
	}
	secretName := saInfo.Secrets[0].Name

	secretInfo, err := k8sClient.CoreV1().Secrets(namespace).Get(secretName, meta_v1.GetOptions{})
	if err != nil{
		glog.Errorln(err)
		return nil, err
	}

	secretData := secretInfo.Data
	certificate := string(secretData["ca.crt"])
	certificate= base64.StdEncoding.EncodeToString([]byte(certificate))
	server := apiserverHost
	token := string(secretData["token"])
	user := string(secretData["namespace"])

	return &Config{Certificate:certificate, Server:server, Token:token, User:user}, nil
}
