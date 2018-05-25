package store // import "github.com/docker/docker/volume/store"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/volume"
	volumedrivers "github.com/docker/docker/volume/drivers"
	volumetestutils "github.com/docker/docker/volume/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	s, cleanup := setupTest(t)
	defer cleanup()
	s.drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")

	v, err := s.Create("fake1", "fake", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.Name() != "fake1" {
		t.Fatalf("Expected fake1 volume, got %v", v)
	}
	if l, _, _ := s.List(); len(l) != 1 {
		t.Fatalf("Expected 1 volume in the store, got %v: %v", len(l), l)
	}

	if _, err := s.Create("none", "none", nil, nil); err == nil {
		t.Fatalf("Expected unknown driver error, got nil")
	}

	_, err = s.Create("fakeerror", "fake", map[string]string{"error": "create error"}, nil)
	expected := &OpErr{Op: "create", Name: "fakeerror", Err: errors.New("create error")}
	if err != nil && err.Error() != expected.Error() {
		t.Fatalf("Expected create fakeError: create error, got %v", err)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	s, cleanup := setupTest(t)
	defer cleanup()

	s.drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")
	s.drivers.Register(volumetestutils.NewFakeDriver("noop"), "noop")

	// doing string compare here since this error comes directly from the driver
	expected := "no such volume"
	if err := s.Remove(volumetestutils.NoopVolume{}); err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Expected error %q, got %v", expected, err)
	}

	v, err := s.CreateWithRef("fake1", "fake", "fake", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Remove(v); !IsInUse(err) {
		t.Fatalf("Expected ErrVolumeInUse error, got %v", err)
	}
	s.Dereference(v, "fake")
	if err := s.Remove(v); err != nil {
		t.Fatal(err)
	}
	if l, _, _ := s.List(); len(l) != 0 {
		t.Fatalf("Expected 0 volumes in the store, got %v, %v", len(l), l)
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "test-list")
	assert.NilError(t, err)
	defer os.RemoveAll(dir)

	drivers := volumedrivers.NewStore(nil)
	drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")
	drivers.Register(volumetestutils.NewFakeDriver("fake2"), "fake2")

	s, err := New(dir, drivers)
	assert.NilError(t, err)

	if _, err := s.Create("test", "fake", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("test2", "fake2", nil, nil); err != nil {
		t.Fatal(err)
	}

	ls, _, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 2 {
		t.Fatalf("expected 2 volumes, got: %d", len(ls))
	}
	if err := s.Shutdown(); err != nil {
		t.Fatal(err)
	}

	// and again with a new store
	s, err = New(dir, drivers)
	if err != nil {
		t.Fatal(err)
	}
	ls, _, err = s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 2 {
		t.Fatalf("expected 2 volumes, got: %d", len(ls))
	}
}

func TestFilterByDriver(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTest(t)
	defer cleanup()

	s.drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")
	s.drivers.Register(volumetestutils.NewFakeDriver("noop"), "noop")

	if _, err := s.Create("fake1", "fake", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("fake2", "fake", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("fake3", "noop", nil, nil); err != nil {
		t.Fatal(err)
	}

	if l, _ := s.FilterByDriver("fake"); len(l) != 2 {
		t.Fatalf("Expected 2 volumes, got %v, %v", len(l), l)
	}

	if l, _ := s.FilterByDriver("noop"); len(l) != 1 {
		t.Fatalf("Expected 1 volume, got %v, %v", len(l), l)
	}
}

func TestFilterByUsed(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTest(t)
	defer cleanup()

	s.drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")
	s.drivers.Register(volumetestutils.NewFakeDriver("noop"), "noop")

	if _, err := s.CreateWithRef("fake1", "fake", "volReference", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("fake2", "fake", nil, nil); err != nil {
		t.Fatal(err)
	}

	vols, _, err := s.List()
	if err != nil {
		t.Fatal(err)
	}

	dangling := s.FilterByUsed(vols, false)
	if len(dangling) != 1 {
		t.Fatalf("expected 1 dangling volume, got %v", len(dangling))
	}
	if dangling[0].Name() != "fake2" {
		t.Fatalf("expected dangling volume fake2, got %s", dangling[0].Name())
	}

	used := s.FilterByUsed(vols, true)
	if len(used) != 1 {
		t.Fatalf("expected 1 used volume, got %v", len(used))
	}
	if used[0].Name() != "fake1" {
		t.Fatalf("expected used volume fake1, got %s", used[0].Name())
	}
}

func TestDerefMultipleOfSameRef(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTest(t)
	defer cleanup()
	s.drivers.Register(volumetestutils.NewFakeDriver("fake"), "fake")

	v, err := s.CreateWithRef("fake1", "fake", "volReference", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetWithRef("fake1", "fake", "volReference"); err != nil {
		t.Fatal(err)
	}

	s.Dereference(v, "volReference")
	if err := s.Remove(v); err != nil {
		t.Fatal(err)
	}
}

func TestCreateKeepOptsLabelsWhenExistsRemotely(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTest(t)
	defer cleanup()

	vd := volumetestutils.NewFakeDriver("fake")
	s.drivers.Register(vd, "fake")

	// Create a volume in the driver directly
	if _, err := vd.Create("foo", nil); err != nil {
		t.Fatal(err)
	}

	v, err := s.Create("foo", "fake", nil, map[string]string{"hello": "world"})
	if err != nil {
		t.Fatal(err)
	}

	switch dv := v.(type) {
	case volume.DetailedVolume:
		if dv.Labels()["hello"] != "world" {
			t.Fatalf("labels don't match")
		}
	default:
		t.Fatalf("got unexpected type: %T", v)
	}
}

func TestDefererencePluginOnCreateError(t *testing.T) {
	t.Parallel()

	var (
		l   net.Listener
		err error
	)

	for i := 32768; l == nil && i < 40000; i++ {
		l, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", i))
	}
	if l == nil {
		t.Fatalf("could not create listener: %v", err)
	}
	defer l.Close()

	s, cleanup := setupTest(t)
	defer cleanup()

	d := volumetestutils.NewFakeDriver("TestDefererencePluginOnCreateError")
	p, err := volumetestutils.MakeFakePlugin(d, l)
	if err != nil {
		t.Fatal(err)
	}

	pg := volumetestutils.NewFakePluginGetter(p)
	s.drivers = volumedrivers.NewStore(pg)

	// create a good volume so we have a plugin reference
	_, err = s.Create("fake1", d.Name(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Now create another one expecting an error
	_, err = s.Create("fake2", d.Name(), map[string]string{"error": "some error"}, nil)
	if err == nil || !strings.Contains(err.Error(), "some error") {
		t.Fatalf("expected an error on create: %v", err)
	}

	// There should be only 1 plugin reference
	if refs := volumetestutils.FakeRefs(p); refs != 1 {
		t.Fatalf("expected 1 plugin reference, got: %d", refs)
	}
}

func TestRefDerefRemove(t *testing.T) {
	t.Parallel()

	driverName := "test-ref-deref-remove"
	s, cleanup := setupTest(t)
	defer cleanup()
	s.drivers.Register(volumetestutils.NewFakeDriver(driverName), driverName)

	v, err := s.CreateWithRef("test", driverName, "test-ref", nil, nil)
	assert.NilError(t, err)

	err = s.Remove(v)
	assert.Assert(t, is.ErrorContains(err, ""))
	assert.Equal(t, errVolumeInUse, err.(*OpErr).Err)

	s.Dereference(v, "test-ref")
	err = s.Remove(v)
	assert.NilError(t, err)
}

func TestGet(t *testing.T) {
	t.Parallel()

	driverName := "test-get"
	s, cleanup := setupTest(t)
	defer cleanup()
	s.drivers.Register(volumetestutils.NewFakeDriver(driverName), driverName)

	_, err := s.Get("not-exist")
	assert.Assert(t, is.ErrorContains(err, ""))
	assert.Equal(t, errNoSuchVolume, err.(*OpErr).Err)

	v1, err := s.Create("test", driverName, nil, map[string]string{"a": "1"})
	assert.NilError(t, err)

	v2, err := s.Get("test")
	assert.NilError(t, err)
	assert.DeepEqual(t, v1, v2, cmpVolume)

	dv := v2.(volume.DetailedVolume)
	assert.Equal(t, "1", dv.Labels()["a"])

	err = s.Remove(v1)
	assert.NilError(t, err)
}

func TestGetWithRef(t *testing.T) {
	t.Parallel()

	driverName := "test-get-with-ref"
	s, cleanup := setupTest(t)
	defer cleanup()
	s.drivers.Register(volumetestutils.NewFakeDriver(driverName), driverName)

	_, err := s.GetWithRef("not-exist", driverName, "test-ref")
	assert.Assert(t, is.ErrorContains(err, ""))

	v1, err := s.Create("test", driverName, nil, map[string]string{"a": "1"})
	assert.NilError(t, err)

	v2, err := s.GetWithRef("test", driverName, "test-ref")
	assert.NilError(t, err)
	assert.DeepEqual(t, v1, v2, cmpVolume)

	err = s.Remove(v2)
	assert.Assert(t, is.ErrorContains(err, ""))
	assert.Equal(t, errVolumeInUse, err.(*OpErr).Err)

	s.Dereference(v2, "test-ref")
	err = s.Remove(v2)
	assert.NilError(t, err)
}

var cmpVolume = cmp.AllowUnexported(volumetestutils.FakeVolume{}, volumeWrapper{})

func setupTest(t *testing.T) (*VolumeStore, func()) {
	t.Helper()

	dirName := strings.Replace(t.Name(), string(os.PathSeparator), "_", -1)
	dir, err := ioutil.TempDir("", dirName)
	assert.NilError(t, err)

	cleanup := func() {
		err := os.RemoveAll(dir)
		assert.Check(t, err)
	}

	s, err := New(dir, volumedrivers.NewStore(nil))
	assert.Check(t, err)
	return s, func() {
		s.Shutdown()
		cleanup()
	}
}
