package registry // import "github.com/docker/docker/registry"

import (
	"os"
	"testing"

	"github.com/gotestyourself/gotestyourself/skip"
)

func TestLookupV1Endpoints(t *testing.T) {
	skip.If(t, os.Getuid() != 0, "skipping test that requires root")
	s, err := NewService(ServiceOptions{})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		hostname    string
		expectedLen int
	}{
		{"example.com", 1},
		{DefaultNamespace, 0},
		{DefaultV2Registry.Host, 0},
		{IndexHostname, 0},
	}

	for _, c := range cases {
		if ret, err := s.lookupV1Endpoints(c.hostname); err != nil || len(ret) != c.expectedLen {
			t.Errorf("lookupV1Endpoints(`"+c.hostname+"`) returned %+v and %+v", ret, err)
		}
	}
}
