package handshake

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

type transportParameterID uint16

const (
	originalConnectionIDParameterID           transportParameterID = 0x0
	idleTimeoutParameterID                    transportParameterID = 0x1
	statelessResetTokenParameterID            transportParameterID = 0x2
	maxPacketSizeParameterID                  transportParameterID = 0x3
	initialMaxDataParameterID                 transportParameterID = 0x4
	initialMaxStreamDataBidiLocalParameterID  transportParameterID = 0x5
	initialMaxStreamDataBidiRemoteParameterID transportParameterID = 0x6
	initialMaxStreamDataUniParameterID        transportParameterID = 0x7
	initialMaxStreamsBidiParameterID          transportParameterID = 0x8
	initialMaxStreamsUniParameterID           transportParameterID = 0x9
	ackDelayExponentParameterID               transportParameterID = 0xa
	disableMigrationParameterID               transportParameterID = 0xc
)

// TransportParameters are parameters sent to the peer during the handshake
type TransportParameters struct {
	InitialMaxStreamDataBidiLocal  protocol.ByteCount
	InitialMaxStreamDataBidiRemote protocol.ByteCount
	InitialMaxStreamDataUni        protocol.ByteCount
	InitialMaxData                 protocol.ByteCount

	AckDelayExponent uint8

	MaxPacketSize protocol.ByteCount

	MaxUniStreams  uint64
	MaxBidiStreams uint64

	IdleTimeout      time.Duration
	DisableMigration bool

	StatelessResetToken  *[16]byte
	OriginalConnectionID protocol.ConnectionID
}

// Unmarshal the transport parameters
func (p *TransportParameters) Unmarshal(data []byte, sentBy protocol.Perspective) error {
	if len(data) < 2 {
		return errors.New("transport parameter data too short")
	}
	length := binary.BigEndian.Uint16(data[:2])
	if len(data)-2 < int(length) {
		return fmt.Errorf("expected transport parameters to be %d bytes long, have %d", length, len(data)-2)
	}

	// needed to check that every parameter is only sent at most once
	var parameterIDs []transportParameterID

	var readAckDelayExponent bool

	r := bytes.NewReader(data[2:])
	for r.Len() >= 4 {
		paramIDInt, _ := utils.BigEndian.ReadUint16(r)
		paramID := transportParameterID(paramIDInt)
		paramLen, _ := utils.BigEndian.ReadUint16(r)
		parameterIDs = append(parameterIDs, paramID)
		switch paramID {
		case ackDelayExponentParameterID:
			readAckDelayExponent = true
			fallthrough
		case initialMaxStreamDataBidiLocalParameterID,
			initialMaxStreamDataBidiRemoteParameterID,
			initialMaxStreamDataUniParameterID,
			initialMaxDataParameterID,
			initialMaxStreamsBidiParameterID,
			initialMaxStreamsUniParameterID,
			idleTimeoutParameterID,
			maxPacketSizeParameterID:
			if err := p.readNumericTransportParameter(r, paramID, int(paramLen)); err != nil {
				return err
			}
		default:
			if r.Len() < int(paramLen) {
				return fmt.Errorf("remaining length (%d) smaller than parameter length (%d)", r.Len(), paramLen)
			}
			switch paramID {
			case disableMigrationParameterID:
				if paramLen != 0 {
					return fmt.Errorf("wrong length for disable_migration: %d (expected empty)", paramLen)
				}
				p.DisableMigration = true
			case statelessResetTokenParameterID:
				if sentBy == protocol.PerspectiveClient {
					return errors.New("client sent a stateless_reset_token")
				}
				if paramLen != 16 {
					return fmt.Errorf("wrong length for stateless_reset_token: %d (expected 16)", paramLen)
				}
				var token [16]byte
				r.Read(token[:])
				p.StatelessResetToken = &token
			case originalConnectionIDParameterID:
				if sentBy == protocol.PerspectiveClient {
					return errors.New("client sent an original_connection_id")
				}
				p.OriginalConnectionID, _ = protocol.ReadConnectionID(r, int(paramLen))
			default:
				r.Seek(int64(paramLen), io.SeekCurrent)
			}
		}
	}

	if !readAckDelayExponent {
		p.AckDelayExponent = protocol.DefaultAckDelayExponent
	}

	// check that every transport parameter was sent at most once
	sort.Slice(parameterIDs, func(i, j int) bool { return parameterIDs[i] < parameterIDs[j] })
	for i := 0; i < len(parameterIDs)-1; i++ {
		if parameterIDs[i] == parameterIDs[i+1] {
			return fmt.Errorf("received duplicate transport parameter %#x", parameterIDs[i])
		}
	}

	if r.Len() != 0 {
		return fmt.Errorf("should have read all data. Still have %d bytes", r.Len())
	}
	return nil
}

func (p *TransportParameters) readNumericTransportParameter(
	r *bytes.Reader,
	paramID transportParameterID,
	expectedLen int,
) error {
	remainingLen := r.Len()
	val, err := utils.ReadVarInt(r)
	if err != nil {
		return fmt.Errorf("error while reading transport parameter %d: %s", paramID, err)
	}
	if remainingLen-r.Len() != expectedLen {
		return fmt.Errorf("inconsistent transport parameter length for %d", paramID)
	}
	switch paramID {
	case initialMaxStreamDataBidiLocalParameterID:
		p.InitialMaxStreamDataBidiLocal = protocol.ByteCount(val)
	case initialMaxStreamDataBidiRemoteParameterID:
		p.InitialMaxStreamDataBidiRemote = protocol.ByteCount(val)
	case initialMaxStreamDataUniParameterID:
		p.InitialMaxStreamDataUni = protocol.ByteCount(val)
	case initialMaxDataParameterID:
		p.InitialMaxData = protocol.ByteCount(val)
	case initialMaxStreamsBidiParameterID:
		p.MaxBidiStreams = val
	case initialMaxStreamsUniParameterID:
		p.MaxUniStreams = val
	case idleTimeoutParameterID:
		p.IdleTimeout = utils.MaxDuration(protocol.MinRemoteIdleTimeout, time.Duration(val)*time.Millisecond)
	case maxPacketSizeParameterID:
		if val < 1200 {
			return fmt.Errorf("invalid value for max_packet_size: %d (minimum 1200)", val)
		}
		p.MaxPacketSize = protocol.ByteCount(val)
	case ackDelayExponentParameterID:
		if val > protocol.MaxAckDelayExponent {
			return fmt.Errorf("invalid value for ack_delay_exponent: %d (maximum %d)", val, protocol.MaxAckDelayExponent)
		}
		p.AckDelayExponent = uint8(val)
	default:
		return fmt.Errorf("TransportParameter BUG: transport parameter %d not found", paramID)
	}
	return nil
}

// Marshal the transport parameters
func (p *TransportParameters) Marshal() []byte {
	b := &bytes.Buffer{}
	b.Write([]byte{0, 0}) // length. Will be replaced later

	// initial_max_stream_data_bidi_local
	utils.BigEndian.WriteUint16(b, uint16(initialMaxStreamDataBidiLocalParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(p.InitialMaxStreamDataBidiLocal))))
	utils.WriteVarInt(b, uint64(p.InitialMaxStreamDataBidiLocal))
	// initial_max_stream_data_bidi_remote
	utils.BigEndian.WriteUint16(b, uint16(initialMaxStreamDataBidiRemoteParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(p.InitialMaxStreamDataBidiRemote))))
	utils.WriteVarInt(b, uint64(p.InitialMaxStreamDataBidiRemote))
	// initial_max_stream_data_uni
	utils.BigEndian.WriteUint16(b, uint16(initialMaxStreamDataUniParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(p.InitialMaxStreamDataUni))))
	utils.WriteVarInt(b, uint64(p.InitialMaxStreamDataUni))
	// initial_max_data
	utils.BigEndian.WriteUint16(b, uint16(initialMaxDataParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(p.InitialMaxData))))
	utils.WriteVarInt(b, uint64(p.InitialMaxData))
	// initial_max_bidi_streams
	utils.BigEndian.WriteUint16(b, uint16(initialMaxStreamsBidiParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(p.MaxBidiStreams)))
	utils.WriteVarInt(b, p.MaxBidiStreams)
	// initial_max_uni_streams
	utils.BigEndian.WriteUint16(b, uint16(initialMaxStreamsUniParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(p.MaxUniStreams)))
	utils.WriteVarInt(b, p.MaxUniStreams)
	// idle_timeout
	idleTimeout := uint64(p.IdleTimeout / time.Millisecond)
	utils.BigEndian.WriteUint16(b, uint16(idleTimeoutParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(idleTimeout)))
	utils.WriteVarInt(b, idleTimeout)
	// max_packet_size
	utils.BigEndian.WriteUint16(b, uint16(maxPacketSizeParameterID))
	utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(protocol.MaxReceivePacketSize))))
	utils.WriteVarInt(b, uint64(protocol.MaxReceivePacketSize))
	// ack_delay_exponent
	// Only send it if is different from the default value.
	if p.AckDelayExponent != protocol.DefaultAckDelayExponent {
		utils.BigEndian.WriteUint16(b, uint16(ackDelayExponentParameterID))
		utils.BigEndian.WriteUint16(b, uint16(utils.VarIntLen(uint64(p.AckDelayExponent))))
		utils.WriteVarInt(b, uint64(p.AckDelayExponent))
	}
	// disable_migration
	if p.DisableMigration {
		utils.BigEndian.WriteUint16(b, uint16(disableMigrationParameterID))
		utils.BigEndian.WriteUint16(b, 0)
	}
	if p.StatelessResetToken != nil {
		utils.BigEndian.WriteUint16(b, uint16(statelessResetTokenParameterID))
		utils.BigEndian.WriteUint16(b, 16)
		b.Write(p.StatelessResetToken[:])
	}
	// original_connection_id
	if p.OriginalConnectionID.Len() > 0 {
		utils.BigEndian.WriteUint16(b, uint16(originalConnectionIDParameterID))
		utils.BigEndian.WriteUint16(b, uint16(p.OriginalConnectionID.Len()))
		b.Write(p.OriginalConnectionID.Bytes())
	}

	data := b.Bytes()
	binary.BigEndian.PutUint16(data[:2], uint16(b.Len()-2))
	return data
}

// String returns a string representation, intended for logging.
func (p *TransportParameters) String() string {
	logString := "&handshake.TransportParameters{OriginalConnectionID: %s, InitialMaxStreamDataBidiLocal: %#x, InitialMaxStreamDataBidiRemote: %#x, InitialMaxStreamDataUni: %#x, InitialMaxData: %#x, MaxBidiStreams: %d, MaxUniStreams: %d, IdleTimeout: %s, AckDelayExponent: %d"
	logParams := []interface{}{p.OriginalConnectionID, p.InitialMaxStreamDataBidiLocal, p.InitialMaxStreamDataBidiRemote, p.InitialMaxStreamDataUni, p.InitialMaxData, p.MaxBidiStreams, p.MaxUniStreams, p.IdleTimeout, p.AckDelayExponent}
	if p.StatelessResetToken != nil { // the client never sends a stateless reset token
		logString += ", StatelessResetToken: %#x"
		logParams = append(logParams, *p.StatelessResetToken)
	}
	logString += "}"
	return fmt.Sprintf(logString, logParams...)
}
