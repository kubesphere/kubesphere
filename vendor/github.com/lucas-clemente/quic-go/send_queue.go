package quic

type sendQueue struct {
	queue     chan *packedPacket
	closeChan chan struct{}
	conn      connection
}

func newSendQueue(conn connection) *sendQueue {
	s := &sendQueue{
		conn:      conn,
		closeChan: make(chan struct{}),
		queue:     make(chan *packedPacket, 1),
	}
	return s
}

func (h *sendQueue) Send(p *packedPacket) {
	h.queue <- p
}

func (h *sendQueue) Run() error {
	var p *packedPacket
	for {
		select {
		case <-h.closeChan:
			return nil
		case p = <-h.queue:
		}
		if err := h.conn.Write(p.raw); err != nil {
			return err
		}
		p.buffer.Release()
	}
}

func (h *sendQueue) Close() {
	close(h.closeChan)
}
