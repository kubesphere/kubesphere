package wire

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

// ParseConnectionID parses the destination connection ID of a packet.
// It uses the data slice for the connection ID.
// That means that the connection ID must not be used after the packet buffer is released.
func ParseConnectionID(data []byte, shortHeaderConnIDLen int) (protocol.ConnectionID, error) {
	if len(data) == 0 {
		return nil, io.EOF
	}
	isLongHeader := data[0]&0x80 > 0
	if !isLongHeader {
		if len(data) < shortHeaderConnIDLen+1 {
			return nil, io.EOF
		}
		return protocol.ConnectionID(data[1 : 1+shortHeaderConnIDLen]), nil
	}
	if len(data) < 6 {
		return nil, io.EOF
	}
	destConnIDLen, _ := decodeConnIDLen(data[5])
	if len(data) < 6+destConnIDLen {
		return nil, io.EOF
	}
	return protocol.ConnectionID(data[6 : 6+destConnIDLen]), nil
}

// IsVersionNegotiationPacket says if this is a version negotiation packet
func IsVersionNegotiationPacket(b []byte) bool {
	if len(b) < 5 {
		return false
	}
	return b[0]&0x80 > 0 && b[1] == 0 && b[2] == 0 && b[3] == 0 && b[4] == 0
}

var errUnsupportedVersion = errors.New("unsupported version")

// The Header is the version independent part of the header
type Header struct {
	Version          protocol.VersionNumber
	SrcConnectionID  protocol.ConnectionID
	DestConnectionID protocol.ConnectionID

	IsLongHeader bool
	Type         protocol.PacketType
	Length       protocol.ByteCount

	Token                []byte
	SupportedVersions    []protocol.VersionNumber // sent in a Version Negotiation Packet
	OrigDestConnectionID protocol.ConnectionID    // sent in the Retry packet

	typeByte  byte
	parsedLen protocol.ByteCount // how many bytes were read while parsing this header
}

// ParsePacket parses a packet.
// If the packet has a long header, the packet is cut according to the length field.
// If we understand the version, the packet is header up unto the packet number.
// Otherwise, only the invariant part of the header is parsed.
func ParsePacket(data []byte, shortHeaderConnIDLen int) (*Header, []byte /* packet data */, []byte /* rest */, error) {
	hdr, err := parseHeader(bytes.NewReader(data), shortHeaderConnIDLen)
	if err != nil {
		if err == errUnsupportedVersion {
			return hdr, nil, nil, nil
		}
		return nil, nil, nil, err
	}
	var rest []byte
	if hdr.IsLongHeader {
		if protocol.ByteCount(len(data)) < hdr.ParsedLen()+hdr.Length {
			return nil, nil, nil, fmt.Errorf("packet length (%d bytes) is smaller than the expected length (%d bytes)", len(data)-int(hdr.ParsedLen()), hdr.Length)
		}
		packetLen := int(hdr.ParsedLen() + hdr.Length)
		rest = data[packetLen:]
		data = data[:packetLen]
	}
	return hdr, data, rest, nil
}

// ParseHeader parses the header.
// For short header packets: up to the packet number.
// For long header packets:
// * if we understand the version: up to the packet number
// * if not, only the invariant part of the header
func parseHeader(b *bytes.Reader, shortHeaderConnIDLen int) (*Header, error) {
	startLen := b.Len()
	h, err := parseHeaderImpl(b, shortHeaderConnIDLen)
	if err != nil {
		return h, err
	}
	h.parsedLen = protocol.ByteCount(startLen - b.Len())
	return h, err
}

func parseHeaderImpl(b *bytes.Reader, shortHeaderConnIDLen int) (*Header, error) {
	typeByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}

	h := &Header{
		typeByte:     typeByte,
		IsLongHeader: typeByte&0x80 > 0,
	}

	if !h.IsLongHeader {
		if h.typeByte&0x40 == 0 {
			return nil, errors.New("not a QUIC packet")
		}
		if err := h.parseShortHeader(b, shortHeaderConnIDLen); err != nil {
			return nil, err
		}
		return h, nil
	}
	return h, h.parseLongHeader(b)
}

func (h *Header) parseShortHeader(b *bytes.Reader, shortHeaderConnIDLen int) error {
	var err error
	h.DestConnectionID, err = protocol.ReadConnectionID(b, shortHeaderConnIDLen)
	return err
}

func (h *Header) parseLongHeader(b *bytes.Reader) error {
	v, err := utils.BigEndian.ReadUint32(b)
	if err != nil {
		return err
	}
	h.Version = protocol.VersionNumber(v)
	if h.Version != 0 && h.typeByte&0x40 == 0 {
		return errors.New("not a QUIC packet")
	}
	connIDLenByte, err := b.ReadByte()
	if err != nil {
		return err
	}
	dcil, scil := decodeConnIDLen(connIDLenByte)
	h.DestConnectionID, err = protocol.ReadConnectionID(b, dcil)
	if err != nil {
		return err
	}
	h.SrcConnectionID, err = protocol.ReadConnectionID(b, scil)
	if err != nil {
		return err
	}
	if h.Version == 0 {
		return h.parseVersionNegotiationPacket(b)
	}
	// If we don't understand the version, we have no idea how to interpret the rest of the bytes
	if !protocol.IsSupportedVersion(protocol.SupportedVersions, h.Version) {
		return errUnsupportedVersion
	}

	switch (h.typeByte & 0x30) >> 4 {
	case 0x0:
		h.Type = protocol.PacketTypeInitial
	case 0x1:
		h.Type = protocol.PacketType0RTT
	case 0x2:
		h.Type = protocol.PacketTypeHandshake
	case 0x3:
		h.Type = protocol.PacketTypeRetry
	}

	if h.Type == protocol.PacketTypeRetry {
		odcil := decodeSingleConnIDLen(h.typeByte & 0xf)
		h.OrigDestConnectionID, err = protocol.ReadConnectionID(b, odcil)
		if err != nil {
			return err
		}
		h.Token = make([]byte, b.Len())
		if _, err := io.ReadFull(b, h.Token); err != nil {
			return err
		}
		return nil
	}

	if h.Type == protocol.PacketTypeInitial {
		tokenLen, err := utils.ReadVarInt(b)
		if err != nil {
			return err
		}
		if tokenLen > uint64(b.Len()) {
			return io.EOF
		}
		h.Token = make([]byte, tokenLen)
		if _, err := io.ReadFull(b, h.Token); err != nil {
			return err
		}
	}

	pl, err := utils.ReadVarInt(b)
	if err != nil {
		return err
	}
	h.Length = protocol.ByteCount(pl)
	return nil
}

func (h *Header) parseVersionNegotiationPacket(b *bytes.Reader) error {
	if b.Len() == 0 {
		return errors.New("Version Negoation packet has empty version list")
	}
	if b.Len()%4 != 0 {
		return errors.New("Version Negotation packet has a version list with an invalid length")
	}
	h.SupportedVersions = make([]protocol.VersionNumber, b.Len()/4)
	for i := 0; b.Len() > 0; i++ {
		v, err := utils.BigEndian.ReadUint32(b)
		if err != nil {
			return err
		}
		h.SupportedVersions[i] = protocol.VersionNumber(v)
	}
	return nil
}

// ParsedLen returns the number of bytes that were consumed when parsing the header
func (h *Header) ParsedLen() protocol.ByteCount {
	return h.parsedLen
}

// ParseExtended parses the version dependent part of the header.
// The Reader has to be set such that it points to the first byte of the header.
func (h *Header) ParseExtended(b *bytes.Reader, ver protocol.VersionNumber) (*ExtendedHeader, error) {
	return h.toExtendedHeader().parse(b, ver)
}

func (h *Header) toExtendedHeader() *ExtendedHeader {
	return &ExtendedHeader{Header: *h}
}

func decodeConnIDLen(enc byte) (int /*dest conn id len*/, int /*src conn id len*/) {
	return decodeSingleConnIDLen(enc >> 4), decodeSingleConnIDLen(enc & 0xf)
}

func decodeSingleConnIDLen(enc uint8) int {
	if enc == 0 {
		return 0
	}
	return int(enc) + 3
}
