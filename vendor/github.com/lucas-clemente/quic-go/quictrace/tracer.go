package quictrace

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"
	"github.com/lucas-clemente/quic-go/quictrace/pb"
)

type traceEvent struct {
	connID protocol.ConnectionID
	ev     Event
}

// A tracer is used to trace a QUIC connection
type tracer struct {
	eventQueue chan traceEvent

	events map[string] /* conn ID */ []traceEvent
}

var _ Tracer = &tracer{}

// NewTracer creates a new Tracer
func NewTracer() Tracer {
	qt := &tracer{
		eventQueue: make(chan traceEvent, 1<<10),
		events:     make(map[string][]traceEvent),
	}
	go qt.run()
	return qt
}

// Trace traces an event
func (t *tracer) Trace(connID protocol.ConnectionID, ev Event) {
	t.eventQueue <- traceEvent{connID: connID, ev: ev}
}

func (t *tracer) run() {
	for tev := range t.eventQueue {
		key := string(tev.connID)
		if _, ok := t.events[key]; !ok {
			t.events[key] = make([]traceEvent, 0, 10*1<<10)
		}
		t.events[key] = append(t.events[key], tev)
	}
}

func (t *tracer) GetAllTraces() map[string][]byte {
	traces := make(map[string][]byte)
	for connID := range t.events {
		data, err := t.emitByConnIDAsString(connID)
		if err != nil {
			panic(err)
		}
		traces[connID] = data
	}
	return traces
}

// Emit emits the serialized protobuf that will be consumed by quic-trace
func (t *tracer) Emit(connID protocol.ConnectionID) ([]byte, error) {
	return t.emitByConnIDAsString(string(connID))
}

func (t *tracer) emitByConnIDAsString(connID string) ([]byte, error) {
	events, ok := t.events[connID]
	if !ok {
		return nil, fmt.Errorf("no trace found for connection ID %s", connID)
	}
	trace := &pb.Trace{
		DestinationConnectionId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		SourceConnectionId:      []byte{1, 2, 3, 4, 5, 6, 7, 8},
		ProtocolVersion:         []byte{0xff, 0, 0, 19},
		Events:                  make([]*pb.Event, len(events)),
	}
	var startTime time.Time
	for i, ev := range events {
		event := ev.ev
		if i == 0 {
			startTime = event.Time
		}

		packetNumber := uint64(event.PacketNumber)
		packetSize := uint64(event.PacketSize)

		trace.Events[i] = &pb.Event{
			TimeUs:          durationToUs(event.Time.Sub(startTime)),
			EventType:       getEventType(event.EventType),
			PacketSize:      &packetSize,
			PacketNumber:    &packetNumber,
			TransportState:  getTransportState(event.TransportState),
			EncryptionLevel: getEncryptionLevel(event.EncryptionLevel),
			Frames:          getFrames(event.Frames),
		}
	}
	delete(t.events, connID)
	return proto.Marshal(trace)
}

func getEventType(evType EventType) *pb.EventType {
	var t pb.EventType
	switch evType {
	case PacketSent:
		t = pb.EventType_PACKET_SENT
	case PacketReceived:
		t = pb.EventType_PACKET_RECEIVED
	case PacketLost:
		t = pb.EventType_PACKET_LOST
	default:
		panic("unknown event type")
	}
	return &t
}

func getEncryptionLevel(encLevel protocol.EncryptionLevel) *pb.EncryptionLevel {
	enc := pb.EncryptionLevel_ENCRYPTION_UNKNOWN
	switch encLevel {
	case protocol.EncryptionInitial:
		enc = pb.EncryptionLevel_ENCRYPTION_INITIAL
	case protocol.EncryptionHandshake:
		enc = pb.EncryptionLevel_ENCRYPTION_HANDSHAKE
	case protocol.Encryption1RTT:
		enc = pb.EncryptionLevel_ENCRYPTION_1RTT
	}
	return &enc
}

func getFrames(wframes []wire.Frame) []*pb.Frame {
	streamFrameType := pb.FrameType_STREAM
	cryptoFrameType := pb.FrameType_CRYPTO
	ackFrameType := pb.FrameType_ACK
	var frames []*pb.Frame
	for _, frame := range wframes {
		switch f := frame.(type) {
		case *wire.CryptoFrame:
			offset := uint64(f.Offset)
			length := uint64(len(f.Data))
			frames = append(frames, &pb.Frame{
				FrameType: &cryptoFrameType,
				CryptoFrameInfo: &pb.CryptoFrameInfo{
					Offset: &offset,
					Length: &length,
				},
			})
		case *wire.StreamFrame:
			streamID := uint64(f.StreamID)
			offset := uint64(f.Offset)
			length := uint64(f.DataLen())
			frames = append(frames, &pb.Frame{
				FrameType: &streamFrameType,
				StreamFrameInfo: &pb.StreamFrameInfo{
					StreamId: &streamID,
					Offset:   &offset,
					Length:   &length,
				},
			})
		case *wire.AckFrame:
			var ackedPackets []*pb.AckBlock
			for _, ackBlock := range f.AckRanges {
				firstPacket := uint64(ackBlock.Smallest)
				lastPacket := uint64(ackBlock.Largest)
				ackedPackets = append(ackedPackets, &pb.AckBlock{
					FirstPacket: &firstPacket,
					LastPacket:  &lastPacket,
				})
			}
			frames = append(frames, &pb.Frame{
				FrameType: &ackFrameType,
				AckInfo: &pb.AckInfo{
					AckDelayUs:   durationToUs(f.DelayTime),
					AckedPackets: ackedPackets,
				},
			})
		}
	}
	return frames
}

func getTransportState(state *TransportState) *pb.TransportState {
	bytesInFlight := uint64(state.BytesInFlight)
	congestionWindow := uint64(state.CongestionWindow)
	ccs := fmt.Sprintf("InSlowStart: %t, InRecovery: %t", state.InSlowStart, state.InRecovery)
	return &pb.TransportState{
		MinRttUs:               durationToUs(state.MinRTT),
		SmoothedRttUs:          durationToUs(state.SmoothedRTT),
		LastRttUs:              durationToUs(state.LatestRTT),
		InFlightBytes:          &bytesInFlight,
		CwndBytes:              &congestionWindow,
		CongestionControlState: &ccs,
	}
}

func durationToUs(d time.Duration) *uint64 {
	dur := uint64(d / 1000)
	return &dur
}
