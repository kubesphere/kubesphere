package handshake

import (
	"crypto"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/marten-seemann/qtls"
)

var quicVersion1Salt = []byte{0xc3, 0xee, 0xf7, 0x12, 0xc7, 0x2e, 0xbb, 0x5a, 0x11, 0xa7, 0xd2, 0x43, 0x2b, 0xb4, 0x63, 0x65, 0xbe, 0xf9, 0xf5, 0x02}

var initialSuite = &qtls.CipherSuiteTLS13{
	ID:     qtls.TLS_AES_128_GCM_SHA256,
	KeyLen: 16,
	AEAD:   qtls.AEADAESGCMTLS13,
	Hash:   crypto.SHA256,
}

// NewInitialAEAD creates a new AEAD for Initial encryption / decryption.
func NewInitialAEAD(connID protocol.ConnectionID, pers protocol.Perspective) (LongHeaderSealer, LongHeaderOpener) {
	clientSecret, serverSecret := computeSecrets(connID)
	var mySecret, otherSecret []byte
	if pers == protocol.PerspectiveClient {
		mySecret = clientSecret
		otherSecret = serverSecret
	} else {
		mySecret = serverSecret
		otherSecret = clientSecret
	}
	myKey, myIV := computeInitialKeyAndIV(mySecret)
	otherKey, otherIV := computeInitialKeyAndIV(otherSecret)

	encrypter := qtls.AEADAESGCMTLS13(myKey, myIV)
	decrypter := qtls.AEADAESGCMTLS13(otherKey, otherIV)

	return newLongHeaderSealer(encrypter, newHeaderProtector(initialSuite, mySecret, true)),
		newLongHeaderOpener(decrypter, newAESHeaderProtector(initialSuite, otherSecret, true))
}

func computeSecrets(connID protocol.ConnectionID) (clientSecret, serverSecret []byte) {
	initialSecret := qtls.HkdfExtract(crypto.SHA256, connID, quicVersion1Salt)
	clientSecret = qtls.HkdfExpandLabel(crypto.SHA256, initialSecret, []byte{}, "client in", crypto.SHA256.Size())
	serverSecret = qtls.HkdfExpandLabel(crypto.SHA256, initialSecret, []byte{}, "server in", crypto.SHA256.Size())
	return
}

func computeInitialKeyAndIV(secret []byte) (key, iv []byte) {
	key = qtls.HkdfExpandLabel(crypto.SHA256, secret, []byte{}, "quic key", 16)
	iv = qtls.HkdfExpandLabel(crypto.SHA256, secret, []byte{}, "quic iv", 12)
	return
}
