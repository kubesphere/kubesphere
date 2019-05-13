package handshake

import (
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/marten-seemann/qtls"
)

const quicTLSExtensionType = 0xffa5

type extensionHandler struct {
	ourParams  []byte
	paramsChan chan []byte

	perspective protocol.Perspective
}

var _ tlsExtensionHandler = &extensionHandler{}

// newExtensionHandler creates a new extension handler
func newExtensionHandler(params []byte, pers protocol.Perspective) tlsExtensionHandler {
	return &extensionHandler{
		ourParams:   params,
		paramsChan:  make(chan []byte),
		perspective: pers,
	}
}

func (h *extensionHandler) GetExtensions(msgType uint8) []qtls.Extension {
	if (h.perspective == protocol.PerspectiveClient && messageType(msgType) != typeClientHello) ||
		(h.perspective == protocol.PerspectiveServer && messageType(msgType) != typeEncryptedExtensions) {
		return nil
	}
	return []qtls.Extension{{
		Type: quicTLSExtensionType,
		Data: h.ourParams,
	}}
}

func (h *extensionHandler) ReceivedExtensions(msgType uint8, exts []qtls.Extension) {
	if (h.perspective == protocol.PerspectiveClient && messageType(msgType) != typeEncryptedExtensions) ||
		(h.perspective == protocol.PerspectiveServer && messageType(msgType) != typeClientHello) {
		return
	}

	var data []byte
	for _, ext := range exts {
		if ext.Type == quicTLSExtensionType {
			data = ext.Data
			break
		}
	}

	h.paramsChan <- data
}

func (h *extensionHandler) TransportParameters() <-chan []byte {
	return h.paramsChan
}
