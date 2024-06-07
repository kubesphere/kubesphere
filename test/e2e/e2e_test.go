package e2e

import (
	"os"
	"testing"

	"kubesphere.io/kubesphere/test/e2e/framework"
)

func TestMain(m *testing.M) {
	framework.ParseFlags()
	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
