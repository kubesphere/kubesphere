package sockjs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (h *handler) jsonp(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("content-type", "application/javascript; charset=UTF-8")

	req.ParseForm()
	callback := req.Form.Get("c")
	if callback == "" {
		http.Error(rw, `"callback" parameter required`, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.(http.Flusher).Flush()

	sess, _ := h.sessionByRequest(req)
	recv := newHTTPReceiver(rw, 1, &jsonpFrameWriter{callback})
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

func (h *handler) jsonpSend(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	var data io.Reader
	data = req.Body

	formReader := strings.NewReader(req.PostFormValue("d"))
	if formReader.Len() != 0 {
		data = formReader
	}
	if data == nil {
		http.Error(rw, "Payload expected.", http.StatusInternalServerError)
		return
	}
	var messages []string
	err := json.NewDecoder(data).Decode(&messages)
	if err == io.EOF {
		http.Error(rw, "Payload expected.", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(rw, "Broken JSON encoding.", http.StatusInternalServerError)
		return
	}
	sessionID, _ := h.parseSessionID(req.URL)
	h.sessionsMux.Lock()
	defer h.sessionsMux.Unlock()
	if sess, ok := h.sessions[sessionID]; !ok {
		http.NotFound(rw, req)
	} else {
		_ = sess.accept(messages...) // TODO(igm) reponse with http.StatusInternalServerError in case of err?
		rw.Header().Set("content-type", "text/plain; charset=UTF-8")
		rw.Write([]byte("ok"))
	}
}

type jsonpFrameWriter struct {
	callback string
}

func (j *jsonpFrameWriter) write(w io.Writer, frame string) (int, error) {
	return fmt.Fprintf(w, "%s(%s);\r\n", j.callback, quote(frame))
}
