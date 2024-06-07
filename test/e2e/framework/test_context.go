package framework

import (
	"flag"
	"os"

	"kubesphere.io/kubesphere/test/e2e/constant"
)

type TestContextType struct {
	Host         string
	InMemoryTest bool
	Username     string
	Password     string
}

func registerFlags(t *TestContextType) {
	flag.BoolVar(&t.InMemoryTest, "in-memory-test", false,
		"Whether KubeSphere controllers and APIServer be started in memory.")
	flag.StringVar(&t.Host, "ks-apiserver", os.Getenv("KS_APISERVER"),
		"KubeSphere API Server IP/DNS")
	flag.StringVar(&t.Username, "username", os.Getenv("KS_USERNAME"),
		"Username to login to KubeSphere API Server")
	flag.StringVar(&t.Password, "password", os.Getenv("KS_PASSWORD"),
		"Password to login to KubeSphere API Server")
}

var TestContext = &TestContextType{}

func setDefaultValue(t *TestContextType) {
	if t.Host == "" {
		t.Host = constant.LocalAPIServer
	}
	if t.Username == "" {
		t.Username = constant.DefaultAdminUser
	}

	if t.Password == "" {
		t.Password = constant.DefaultPassword
	}
}

func ParseFlags() {
	registerFlags(TestContext)
	flag.Parse()
	setDefaultValue(TestContext)
}
