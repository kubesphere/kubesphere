package build

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/integration/internal/container"
	"github.com/docker/docker/internal/test/fakecontext"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/skip"
)

func TestBuildSquashParent(t *testing.T) {
	skip.If(t, !testEnv.DaemonInfo.ExperimentalBuild)

	client := testEnv.APIClient()

	dockerfile := `
		FROM busybox
		RUN echo hello > /hello
		RUN echo world >> /hello
		RUN echo hello > /remove_me
		ENV HELLO world
		RUN rm /remove_me
		`

	// build and get the ID that we can use later for history comparison
	ctx := context.Background()
	source := fakecontext.New(t, "", fakecontext.WithDockerfile(dockerfile))
	defer source.Close()

	name := "test"
	resp, err := client.ImageBuild(ctx,
		source.AsTarReader(t),
		types.ImageBuildOptions{
			Remove:      true,
			ForceRemove: true,
			Tags:        []string{name},
		})
	assert.NilError(t, err)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	assert.NilError(t, err)

	inspect, _, err := client.ImageInspectWithRaw(ctx, name)
	assert.NilError(t, err)
	origID := inspect.ID

	// build with squash
	resp, err = client.ImageBuild(ctx,
		source.AsTarReader(t),
		types.ImageBuildOptions{
			Remove:      true,
			ForceRemove: true,
			Squash:      true,
			Tags:        []string{name},
		})
	assert.NilError(t, err)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	assert.NilError(t, err)

	cid := container.Run(t, ctx, client,
		container.WithImage(name),
		container.WithCmd("/bin/sh", "-c", "cat /hello"),
	)
	reader, err := client.ContainerLogs(ctx, cid, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	assert.NilError(t, err)

	actualStdout := new(bytes.Buffer)
	actualStderr := ioutil.Discard
	_, err = stdcopy.StdCopy(actualStdout, actualStderr, reader)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(strings.TrimSpace(actualStdout.String()), "hello\nworld"))

	container.Run(t, ctx, client,
		container.WithImage(name),
		container.WithCmd("/bin/sh", "-c", "[ ! -f /remove_me ]"),
	)
	container.Run(t, ctx, client,
		container.WithImage(name),
		container.WithCmd("/bin/sh", "-c", `[ "$(echo $HELLO)" == "world" ]`),
	)

	origHistory, err := client.ImageHistory(ctx, origID)
	assert.NilError(t, err)
	testHistory, err := client.ImageHistory(ctx, name)
	assert.NilError(t, err)

	inspect, _, err = client.ImageInspectWithRaw(ctx, name)
	assert.NilError(t, err)
	assert.Check(t, is.Len(testHistory, len(origHistory)+1))
	assert.Check(t, is.Len(inspect.RootFS.Layers, 2))
}
