package sockjs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type frameWriter interface {
	write(writer io.Writer, frame string) (int, error)
}

type httpReceiverState int

const (
	stateHTTPReceiverActive httpReceiverState = iota
	stateHTTPReceiverClosed
)

type httpReceiver struct {
	sync.Mutex
	state httpReceiverState

	frameWriter         frameWriter
	rw                  http.ResponseWriter
	maxResponseSize     uint32
	currentResponseSize uint32
	doneCh              chan struct{}
	interruptCh         chan struct{}
}

func newHTTPReceiver(rw http.ResponseWriter, maxResponse uint32, frameWriter frameWriter) *httpReceiver {
	recv := &httpReceiver{
		rw:              rw,
		frameWriter:     frameWriter,
		maxResponseSize: maxResponse,
		doneCh:          make(chan struct{}),
		interruptCh:     make(chan struct{}),
	}
	if closeNotifier, ok := rw.(http.CloseNotifier); ok {
		// if supported check for close notifications from http.RW
		closeNotifyCh := closeNotifier.CloseNotify()
		go func() {
			select {
			case <-closeNotifyCh:
				recv.Lock()
				defer recv.Unlock()
				if recv.state < stateHTTPReceiverClosed {
					recv.state = stateHTTPReceiverClosed
					close(recv.interruptCh)
				}
			case <-recv.doneCh:
				// ok, no action needed here, receiver closed in correct way
				// just finish the routine
			}
		}()
	}
	return recv
}

func (recv *httpReceiver) sendBulk(messages ...string) {
	if len(messages) > 0 {
		recv.sendFrame(fmt.Sprintf("a[%s]",
			strings.Join(
				transform(messages, quote),
				",",
			),
		))
	}
}

func (recv *httpReceiver) sendFrame(value string) {
	recv.Lock()
	defer recv.Unlock()

	if recv.state == stateHTTPReceiverActive {
		// TODO(igm) check err, possibly act as if interrupted
		n, _ := recv.frameWriter.write(recv.rw, value)
		recv.currentResponseSize += uint32(n)
		if recv.currentResponseSize >= recv.maxResponseSize {
			recv.state = stateHTTPReceiverClosed
			close(recv.doneCh)
		} else {
			recv.rw.(http.Flusher).Flush()
		}
	}
}

func (recv *httpReceiver) doneNotify() <-chan struct{}        { return recv.doneCh }
func (recv *httpReceiver) interruptedNotify() <-chan struct{} { return recv.interruptCh }
func (recv *httpReceiver) close() {
	recv.Lock()
	defer recv.Unlock()
	if recv.state < stateHTTPReceiverClosed {
		recv.state = stateHTTPReceiverClosed
		close(recv.doneCh)
	}
}
func (recv *httpReceiver) canSend() bool {
	recv.Lock()
	defer recv.Unlock()
	return recv.state != stateHTTPReceiverClosed
}
