package sockjs

import (
	"fmt"
	"io"
	"net/http"
)

func (h *handler) eventSource(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("content-type", "text/event-stream; charset=UTF-8")
	fmt.Fprintf(rw, "\r\n")
	rw.(http.Flusher).Flush()

	recv := newHTTPReceiver(rw, h.options.ResponseLimit, new(eventSourceFrameWriter))
	sess, _ := h.sessionByRequest(req)
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

type eventSourceFrameWriter struct{}

func (*eventSourceFrameWriter) write(w io.Writer, frame string) (int, error) {
	return fmt.Fprintf(w, "data: %s\r\n\r\n", frame)
}
