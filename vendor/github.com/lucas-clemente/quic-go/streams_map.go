package quic

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/lucas-clemente/quic-go/internal/flowcontrol"
	"github.com/lucas-clemente/quic-go/internal/handshake"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type streamError struct {
	message string
	nums    []protocol.StreamNum
}

func (e streamError) Error() string {
	return e.message
}

func convertStreamError(err error, stype protocol.StreamType, pers protocol.Perspective) error {
	strError, ok := err.(streamError)
	if !ok {
		return err
	}
	ids := make([]interface{}, len(strError.nums))
	for i, num := range strError.nums {
		ids[i] = num.StreamID(stype, pers)
	}
	return fmt.Errorf(strError.Error(), ids...)
}

type streamOpenErr struct{ error }

var _ net.Error = &streamOpenErr{}

func (e streamOpenErr) Temporary() bool { return e.error == errTooManyOpenStreams }
func (streamOpenErr) Timeout() bool     { return false }

// errTooManyOpenStreams is used internally by the outgoing streams maps.
var errTooManyOpenStreams = errors.New("too many open streams")

type streamsMap struct {
	perspective protocol.Perspective

	sender            streamSender
	newFlowController func(protocol.StreamID) flowcontrol.StreamFlowController

	outgoingBidiStreams *outgoingBidiStreamsMap
	outgoingUniStreams  *outgoingUniStreamsMap
	incomingBidiStreams *incomingBidiStreamsMap
	incomingUniStreams  *incomingUniStreamsMap
}

var _ streamManager = &streamsMap{}

func newStreamsMap(
	sender streamSender,
	newFlowController func(protocol.StreamID) flowcontrol.StreamFlowController,
	maxIncomingBidiStreams uint64,
	maxIncomingUniStreams uint64,
	perspective protocol.Perspective,
	version protocol.VersionNumber,
) streamManager {
	m := &streamsMap{
		perspective:       perspective,
		newFlowController: newFlowController,
		sender:            sender,
	}
	m.outgoingBidiStreams = newOutgoingBidiStreamsMap(
		func(num protocol.StreamNum) streamI {
			id := num.StreamID(protocol.StreamTypeBidi, perspective)
			return newStream(id, m.sender, m.newFlowController(id), version)
		},
		sender.queueControlFrame,
	)
	m.incomingBidiStreams = newIncomingBidiStreamsMap(
		func(num protocol.StreamNum) streamI {
			id := num.StreamID(protocol.StreamTypeBidi, perspective.Opposite())
			return newStream(id, m.sender, m.newFlowController(id), version)
		},
		maxIncomingBidiStreams,
		sender.queueControlFrame,
	)
	m.outgoingUniStreams = newOutgoingUniStreamsMap(
		func(num protocol.StreamNum) sendStreamI {
			id := num.StreamID(protocol.StreamTypeUni, perspective)
			return newSendStream(id, m.sender, m.newFlowController(id), version)
		},
		sender.queueControlFrame,
	)
	m.incomingUniStreams = newIncomingUniStreamsMap(
		func(num protocol.StreamNum) receiveStreamI {
			id := num.StreamID(protocol.StreamTypeUni, perspective.Opposite())
			return newReceiveStream(id, m.sender, m.newFlowController(id), version)
		},
		maxIncomingUniStreams,
		sender.queueControlFrame,
	)
	return m
}

func (m *streamsMap) OpenStream() (Stream, error) {
	str, err := m.outgoingBidiStreams.OpenStream()
	return str, convertStreamError(err, protocol.StreamTypeBidi, m.perspective)
}

func (m *streamsMap) OpenStreamSync(ctx context.Context) (Stream, error) {
	str, err := m.outgoingBidiStreams.OpenStreamSync(ctx)
	return str, convertStreamError(err, protocol.StreamTypeBidi, m.perspective)
}

func (m *streamsMap) OpenUniStream() (SendStream, error) {
	str, err := m.outgoingUniStreams.OpenStream()
	return str, convertStreamError(err, protocol.StreamTypeBidi, m.perspective)
}

func (m *streamsMap) OpenUniStreamSync(ctx context.Context) (SendStream, error) {
	str, err := m.outgoingUniStreams.OpenStreamSync(ctx)
	return str, convertStreamError(err, protocol.StreamTypeUni, m.perspective)
}

func (m *streamsMap) AcceptStream(ctx context.Context) (Stream, error) {
	str, err := m.incomingBidiStreams.AcceptStream(ctx)
	return str, convertStreamError(err, protocol.StreamTypeBidi, m.perspective.Opposite())
}

func (m *streamsMap) AcceptUniStream(ctx context.Context) (ReceiveStream, error) {
	str, err := m.incomingUniStreams.AcceptStream(ctx)
	return str, convertStreamError(err, protocol.StreamTypeUni, m.perspective.Opposite())
}

func (m *streamsMap) DeleteStream(id protocol.StreamID) error {
	num := id.StreamNum()
	switch id.Type() {
	case protocol.StreamTypeUni:
		if id.InitiatedBy() == m.perspective {
			return convertStreamError(m.outgoingUniStreams.DeleteStream(num), protocol.StreamTypeUni, m.perspective)
		}
		return convertStreamError(m.incomingUniStreams.DeleteStream(num), protocol.StreamTypeUni, m.perspective.Opposite())
	case protocol.StreamTypeBidi:
		if id.InitiatedBy() == m.perspective {
			return convertStreamError(m.outgoingBidiStreams.DeleteStream(num), protocol.StreamTypeBidi, m.perspective)
		}
		return convertStreamError(m.incomingBidiStreams.DeleteStream(num), protocol.StreamTypeBidi, m.perspective.Opposite())
	}
	panic("")
}

func (m *streamsMap) GetOrOpenReceiveStream(id protocol.StreamID) (receiveStreamI, error) {
	str, err := m.getOrOpenReceiveStream(id)
	if err != nil {
		return nil, qerr.Error(qerr.StreamStateError, err.Error())
	}
	return str, nil
}

func (m *streamsMap) getOrOpenReceiveStream(id protocol.StreamID) (receiveStreamI, error) {
	num := id.StreamNum()
	switch id.Type() {
	case protocol.StreamTypeUni:
		if id.InitiatedBy() == m.perspective {
			// an outgoing unidirectional stream is a send stream, not a receive stream
			return nil, fmt.Errorf("peer attempted to open receive stream %d", id)
		}
		str, err := m.incomingUniStreams.GetOrOpenStream(num)
		return str, convertStreamError(err, protocol.StreamTypeUni, m.perspective)
	case protocol.StreamTypeBidi:
		var str receiveStreamI
		var err error
		if id.InitiatedBy() == m.perspective {
			str, err = m.outgoingBidiStreams.GetStream(num)
		} else {
			str, err = m.incomingBidiStreams.GetOrOpenStream(num)
		}
		return str, convertStreamError(err, protocol.StreamTypeBidi, id.InitiatedBy())
	}
	panic("")
}

func (m *streamsMap) GetOrOpenSendStream(id protocol.StreamID) (sendStreamI, error) {
	str, err := m.getOrOpenSendStream(id)
	if err != nil {
		return nil, qerr.Error(qerr.StreamStateError, err.Error())
	}
	return str, nil
}

func (m *streamsMap) getOrOpenSendStream(id protocol.StreamID) (sendStreamI, error) {
	num := id.StreamNum()
	switch id.Type() {
	case protocol.StreamTypeUni:
		if id.InitiatedBy() == m.perspective {
			str, err := m.outgoingUniStreams.GetStream(num)
			return str, convertStreamError(err, protocol.StreamTypeUni, m.perspective)
		}
		// an incoming unidirectional stream is a receive stream, not a send stream
		return nil, fmt.Errorf("peer attempted to open send stream %d", id)
	case protocol.StreamTypeBidi:
		var str sendStreamI
		var err error
		if id.InitiatedBy() == m.perspective {
			str, err = m.outgoingBidiStreams.GetStream(num)
		} else {
			str, err = m.incomingBidiStreams.GetOrOpenStream(num)
		}
		return str, convertStreamError(err, protocol.StreamTypeBidi, id.InitiatedBy())
	}
	panic("")
}

func (m *streamsMap) HandleMaxStreamsFrame(f *wire.MaxStreamsFrame) error {
	switch f.Type {
	case protocol.StreamTypeUni:
		m.outgoingUniStreams.SetMaxStream(f.MaxStreamNum)
	case protocol.StreamTypeBidi:
		m.outgoingBidiStreams.SetMaxStream(f.MaxStreamNum)
	}
	return nil
}

func (m *streamsMap) UpdateLimits(p *handshake.TransportParameters) error {
	if p.MaxBidiStreamNum > protocol.MaxStreamCount ||
		p.MaxUniStreamNum > protocol.MaxStreamCount {
		return qerr.StreamLimitError
	}
	// Max{Uni,Bidi}StreamID returns the highest stream ID that the peer is allowed to open.
	m.outgoingBidiStreams.SetMaxStream(p.MaxBidiStreamNum)
	m.outgoingUniStreams.SetMaxStream(p.MaxUniStreamNum)
	return nil
}

func (m *streamsMap) CloseWithError(err error) {
	m.outgoingBidiStreams.CloseWithError(err)
	m.outgoingUniStreams.CloseWithError(err)
	m.incomingBidiStreams.CloseWithError(err)
	m.incomingUniStreams.CloseWithError(err)
}
