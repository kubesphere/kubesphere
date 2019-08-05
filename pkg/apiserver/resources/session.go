/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package resources

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"kubesphere.io/kubesphere/pkg/controller/watch"
	"time"
)

const (
	// idleSessionTimeout defines duration of being idle before terminating a session.
	idleSessionTimeout = time.Second * 55

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = idleSessionTimeout

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	sessions      = make(map[string]*Session, 0)
	EventReceiver = &WebsocketEventReceiver{}
)

type Session struct {
	id        string
	username  string
	conn      *websocket.Conn
	namespace string
	queue     chan interface{}
}

func (session *Session) String() string {
	return fmt.Sprintf("username: %s,namespace: %s,queue: %d,id: %s", session.username, session.namespace, len(session.queue), session.id)
}

func NewSession(username string, conn *websocket.Conn) *Session {
	session := &Session{conn: conn}
	session.id = uuid.New().String()
	session.username = username
	session.queue = make(chan interface{}, 256)
	return session
}

func (session *Session) subscribe(namespace string) {
	session.namespace = namespace
	sessions[session.id] = session
	go session.readLoop()
	go session.writeLoop()
	glog.V(4).Infoln("session created", session)
}
func (session *Session) close() {
	delete(sessions, session.id)
	session.conn.Close()
	glog.V(4).Infoln("session closed", session)
}

func (session *Session) send(data []byte) bool {
	if session == nil {
		return true
	}

	select {
	case session.queue <- data:
	case <-time.After(idleSessionTimeout):
		glog.Infoln("session.send: timeout", session)
		return false
	}
	return true
}
func (session *Session) readLoop() {
	defer func() {
		session.close()
	}()

	session.conn.SetReadLimit(1024)
	session.conn.SetReadDeadline(time.Now().Add(idleSessionTimeout))
	session.conn.SetPongHandler(func(string) error {
		session.conn.SetReadDeadline(time.Now().Add(idleSessionTimeout))
		return nil
	})

	for {
		_, raw, err := session.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				glog.V(4).Infoln(session, err)
			}
			return
		}
		session.dispatchRaw(raw)
	}
}

func (session *Session) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		session.close()
	}()

	for {
		select {
		case msg, ok := <-session.queue:
			if !ok {
				// Channel closed.
				return
			}
			if err := session.write(websocket.TextMessage, msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					glog.V(4).Infoln(session, err)
				}
				return
			}
		case <-ticker.C:
			if err := session.write(websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					glog.V(4).Infoln(session, err)
				}
				return
			}
		}
	}
}

func (session *Session) write(mt int, msg interface{}) error {
	var bits []byte
	if msg != nil {
		bits = msg.([]byte)
	} else {
		bits = []byte{}
	}
	session.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return session.conn.WriteMessage(mt, bits)
}

func (session *Session) dispatchRaw(raw []byte) {
	if len(raw) == 1 && raw[0] == 0x31 {
		session.send([]byte{0x30})
		return
	}
}

type WebsocketEventReceiver struct {
}

func (eventHandler *WebsocketEventReceiver) HandleEvent(event watch.Event) {
	for _, session := range sessions {
		if session.namespace == event.Namespace {
			data, _ := json.Marshal(event)
			session.send(data)
		}
	}
}
