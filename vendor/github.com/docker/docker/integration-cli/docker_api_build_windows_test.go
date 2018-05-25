// +build windows

package main

import (
	"net/http"

	"github.com/docker/docker/integration-cli/checker"
	"github.com/docker/docker/internal/test/fakecontext"
	"github.com/docker/docker/internal/test/request"
	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

func (s *DockerSuite) TestBuildWithRecycleBin(c *check.C) {
	testRequires(c, DaemonIsWindows)

	dockerfile := "" +
		"FROM " + testEnv.PlatformDefaults.BaseImage + "\n" +
		"RUN md $REcycLE.biN && md missing\n" +
		"RUN dir $Recycle.Bin && exit 1 || exit 0\n" +
		"RUN dir missing\n"

	ctx := fakecontext.New(c, "", fakecontext.WithDockerfile(dockerfile))
	defer ctx.Close()

	res, body, err := request.Post(
		"/build",
		request.RawContent(ctx.AsTarReader(c)),
		request.ContentType("application/x-tar"))

	c.Assert(err, checker.IsNil)
	c.Assert(res.StatusCode, checker.Equals, http.StatusOK)

	out, err := request.ReadBody(body)
	assert.NilError(c, err)
	assert.Check(c, is.Contains(string(out), "Successfully built"))
}
