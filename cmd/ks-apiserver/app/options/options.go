package options

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)

type ServerRunOptions struct {

	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// OpenPitrix api gateway service url
	OpenPitrixAddress string

	// database connection string in MySQL like
	// user:password@tcp(host)/dbname?charset=utf8&parseTime=True
	DatabaseConnectionString string

	// tls cert file
	TlsCertFile string

	// tls private key file
	TlsPrivateKey string

	// host openapi doc
	ApiDoc bool

	// kubeconfig file path
	KubeConfig string
}

func NewServerRunOptions() *ServerRunOptions {
	// create default server run options
	s := ServerRunOptions{
		BindAddress:              "0.0.0.0",
		InsecurePort:             9090,
		SecurePort:               0,
		OpenPitrixAddress:        "openpitrix-api-gateway.openpitrix-system.svc",
		DatabaseConnectionString: "",
		TlsCertFile:              "",
		TlsPrivateKey:            "",
		ApiDoc:                   true,
	}

	return &s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.BindAddress, "bind-address", "0.0.0.0", "server bind address")
	fs.IntVar(&s.InsecurePort, "insecure-port", 9090, "insecure port number")
	fs.IntVar(&s.SecurePort, "secure-port", 0, "secure port number")
	fs.StringVar(&s.OpenPitrixAddress, "openpitrix", "openpitrix-api-gateway.openpitrix-system.svc", "openpitrix api gateway address")
	fs.StringVar(&s.DatabaseConnectionString, "database-connection", "", "database connection string")
	fs.StringVar(&s.TlsCertFile, "tls-cert-file", "", "tls cert file")
	fs.StringVar(&s.TlsPrivateKey, "tls-private-key", "", "tls private key")
	fs.BoolVar(&s.ApiDoc, "api-doc", true, "host OpenAPI doc")
	fs.StringVar(&s.KubeConfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")), "path to kubeconfig file")
}
