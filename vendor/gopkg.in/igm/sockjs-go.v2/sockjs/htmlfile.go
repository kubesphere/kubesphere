package sockjs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

var iframeTemplate = `<!doctype html>
<html><head>
  <meta http-equiv="X-UA-Compatible" content="IE=edge" />
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
</head><body><h2>Don't panic!</h2>
  <script>
    document.domain = document.domain;
    var c = parent.%s;
    c.start();
    function p(d) {c.message(d);};
    window.onload = function() {c.stop();};
  </script>
`

func init() {
	iframeTemplate += strings.Repeat(" ", 1024-len(iframeTemplate)+14)
	iframeTemplate += "\r\n\r\n"
}

func (h *handler) htmlFile(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("content-type", "text/html; charset=UTF-8")

	req.ParseForm()
	callback := req.Form.Get("c")
	if callback == "" {
		http.Error(rw, `"callback" parameter required`, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, iframeTemplate, callback)
	rw.(http.Flusher).Flush()
	sess, _ := h.sessionByRequest(req)
	recv := newHTTPReceiver(rw, h.options.ResponseLimit, new(htmlfileFrameWriter))
	if err := sess.attachReceiver(recv); err != nil {
		recv.sendFrame(cFrame)
		recv.close()
		return
	}
	select {
	case <-recv.doneNotify():
	case <-recv.interruptedNotify():
	}
}

type htmlfileFrameWriter struct{}

func (*htmlfileFrameWriter) write(w io.Writer, frame string) (int, error) {
	return fmt.Fprintf(w, "<script>\np(%s);\n</script>\r\n", quote(frame))
}
