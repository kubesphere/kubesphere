package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/docker/docker/internal/test/request"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/go-check/check"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

func (s *DockerSuite) TestGetContainersAttachWebsocket(c *check.C) {
	testRequires(c, DaemonIsLinux)
	out, _ := dockerCmd(c, "run", "-dit", "busybox", "cat")

	rwc, err := request.SockConn(time.Duration(10*time.Second), daemonHost())
	c.Assert(err, checker.IsNil)

	cleanedContainerID := strings.TrimSpace(out)
	config, err := websocket.NewConfig(
		"/containers/"+cleanedContainerID+"/attach/ws?stream=1&stdin=1&stdout=1&stderr=1",
		"http://localhost",
	)
	c.Assert(err, checker.IsNil)

	ws, err := websocket.NewClient(config, rwc)
	c.Assert(err, checker.IsNil)
	defer ws.Close()

	expected := []byte("hello")
	actual := make([]byte, len(expected))

	outChan := make(chan error)
	go func() {
		_, err := io.ReadFull(ws, actual)
		outChan <- err
		close(outChan)
	}()

	inChan := make(chan error)
	go func() {
		_, err := ws.Write(expected)
		inChan <- err
		close(inChan)
	}()

	select {
	case err := <-inChan:
		c.Assert(err, checker.IsNil)
	case <-time.After(5 * time.Second):
		c.Fatal("Timeout writing to ws")
	}

	select {
	case err := <-outChan:
		c.Assert(err, checker.IsNil)
	case <-time.After(5 * time.Second):
		c.Fatal("Timeout reading from ws")
	}

	c.Assert(actual, checker.DeepEquals, expected, check.Commentf("Websocket didn't return the expected data"))
}

// regression gh14320
func (s *DockerSuite) TestPostContainersAttachContainerNotFound(c *check.C) {
	resp, _, err := request.Post("/containers/doesnotexist/attach")
	c.Assert(err, checker.IsNil)
	// connection will shutdown, err should be "persistent connection closed"
	c.Assert(resp.StatusCode, checker.Equals, http.StatusNotFound)
	content, err := request.ReadBody(resp.Body)
	c.Assert(err, checker.IsNil)
	expected := "No such container: doesnotexist\r\n"
	c.Assert(string(content), checker.Equals, expected)
}

func (s *DockerSuite) TestGetContainersWsAttachContainerNotFound(c *check.C) {
	res, body, err := request.Get("/containers/doesnotexist/attach/ws")
	c.Assert(res.StatusCode, checker.Equals, http.StatusNotFound)
	c.Assert(err, checker.IsNil)
	b, err := request.ReadBody(body)
	c.Assert(err, checker.IsNil)
	expected := "No such container: doesnotexist"
	c.Assert(getErrorMessage(c, b), checker.Contains, expected)
}

func (s *DockerSuite) TestPostContainersAttach(c *check.C) {
	testRequires(c, DaemonIsLinux)

	expectSuccess := func(conn net.Conn, br *bufio.Reader, stream string, tty bool) {
		defer conn.Close()
		expected := []byte("success")
		_, err := conn.Write(expected)
		c.Assert(err, checker.IsNil)

		conn.SetReadDeadline(time.Now().Add(time.Second))
		lenHeader := 0
		if !tty {
			lenHeader = 8
		}
		actual := make([]byte, len(expected)+lenHeader)
		_, err = io.ReadFull(br, actual)
		c.Assert(err, checker.IsNil)
		if !tty {
			fdMap := map[string]byte{
				"stdin":  0,
				"stdout": 1,
				"stderr": 2,
			}
			c.Assert(actual[0], checker.Equals, fdMap[stream])
		}
		c.Assert(actual[lenHeader:], checker.DeepEquals, expected, check.Commentf("Attach didn't return the expected data from %s", stream))
	}

	expectTimeout := func(conn net.Conn, br *bufio.Reader, stream string) {
		defer conn.Close()
		_, err := conn.Write([]byte{'t'})
		c.Assert(err, checker.IsNil)

		conn.SetReadDeadline(time.Now().Add(time.Second))
		actual := make([]byte, 1)
		_, err = io.ReadFull(br, actual)
		opErr, ok := err.(*net.OpError)
		c.Assert(ok, checker.Equals, true, check.Commentf("Error is expected to be *net.OpError, got %v", err))
		c.Assert(opErr.Timeout(), checker.Equals, true, check.Commentf("Read from %s is expected to timeout", stream))
	}

	// Create a container that only emits stdout.
	cid, _ := dockerCmd(c, "run", "-di", "busybox", "cat")
	cid = strings.TrimSpace(cid)
	// Attach to the container's stdout stream.
	conn, br, err := sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stdout=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	// Check if the data from stdout can be received.
	expectSuccess(conn, br, "stdout", false)
	// Attach to the container's stderr stream.
	conn, br, err = sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stderr=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	// Since the container only emits stdout, attaching to stderr should return nothing.
	expectTimeout(conn, br, "stdout")

	// Test the similar functions of the stderr stream.
	cid, _ = dockerCmd(c, "run", "-di", "busybox", "/bin/sh", "-c", "cat >&2")
	cid = strings.TrimSpace(cid)
	conn, br, err = sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stderr=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	expectSuccess(conn, br, "stderr", false)
	conn, br, err = sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stdout=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	expectTimeout(conn, br, "stderr")

	// Test with tty.
	cid, _ = dockerCmd(c, "run", "-dit", "busybox", "/bin/sh", "-c", "cat >&2")
	cid = strings.TrimSpace(cid)
	// Attach to stdout only.
	conn, br, err = sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stdout=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	expectSuccess(conn, br, "stdout", true)

	// Attach without stdout stream.
	conn, br, err = sockRequestHijack("POST", "/containers/"+cid+"/attach?stream=1&stdin=1&stderr=1", nil, "text/plain", daemonHost())
	c.Assert(err, checker.IsNil)
	// Nothing should be received because both the stdout and stderr of the container will be
	// sent to the client as stdout when tty is enabled.
	expectTimeout(conn, br, "stdout")

	// Test the client API
	client, err := client.NewEnvClient()
	c.Assert(err, checker.IsNil)
	defer client.Close()

	cid, _ = dockerCmd(c, "run", "-di", "busybox", "/bin/sh", "-c", "echo hello; cat")
	cid = strings.TrimSpace(cid)

	// Make sure we don't see "hello" if Logs is false
	attachOpts := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   false,
	}

	resp, err := client.ContainerAttach(context.Background(), cid, attachOpts)
	c.Assert(err, checker.IsNil)
	expectSuccess(resp.Conn, resp.Reader, "stdout", false)

	// Make sure we do see "hello" if Logs is true
	attachOpts.Logs = true
	resp, err = client.ContainerAttach(context.Background(), cid, attachOpts)
	c.Assert(err, checker.IsNil)

	defer resp.Conn.Close()
	resp.Conn.SetReadDeadline(time.Now().Add(time.Second))

	_, err = resp.Conn.Write([]byte("success"))
	c.Assert(err, checker.IsNil)

	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
	if err != nil && errors.Cause(err).(net.Error).Timeout() {
		// ignore the timeout error as it is expected
		err = nil
	}
	c.Assert(err, checker.IsNil)
	c.Assert(errBuf.String(), checker.Equals, "")
	c.Assert(outBuf.String(), checker.Equals, "hello\nsuccess")
}

// SockRequestHijack creates a connection to specified host (with method, contenttype, …) and returns a hijacked connection
// and the output as a `bufio.Reader`
func sockRequestHijack(method, endpoint string, data io.Reader, ct string, daemon string, modifiers ...func(*http.Request)) (net.Conn, *bufio.Reader, error) {
	req, client, err := newRequestClient(method, endpoint, data, ct, daemon, modifiers...)
	if err != nil {
		return nil, nil, err
	}

	client.Do(req)
	conn, br := client.Hijack()
	return conn, br, nil
}

// FIXME(vdemeester) httputil.ClientConn is deprecated, use http.Client instead (closer to actual client)
// Deprecated: Use New instead of NewRequestClient
// Deprecated: use request.Do (or Get, Delete, Post) instead
func newRequestClient(method, endpoint string, data io.Reader, ct, daemon string, modifiers ...func(*http.Request)) (*http.Request, *httputil.ClientConn, error) {
	c, err := request.SockConn(time.Duration(10*time.Second), daemon)
	if err != nil {
		return nil, nil, fmt.Errorf("could not dial docker daemon: %v", err)
	}

	client := httputil.NewClientConn(c, nil)

	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("could not create new request: %v", err)
	}

	for _, opt := range modifiers {
		opt(req)
	}

	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	return req, client, nil
}
