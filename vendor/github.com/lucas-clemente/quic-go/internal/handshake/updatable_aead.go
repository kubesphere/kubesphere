package handshake

import (
	"crypto"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/marten-seemann/qtls"
)

// By setting this environment variable, the key update interval can be adjusted.
// This is not needed in production, but useful for integration and interop testing.
// Note that no mattter what value is set, a key update is only initiated once it is
// permitted (i.e. once an ACK for a packet sent at the current key phase has been received).
const keyUpdateEnv = "QUIC_GO_KEY_UPDATE_INTERVAL"

var keyUpdateInterval uint64

func init() {
	setKeyUpdateInterval()
}

func setKeyUpdateInterval() {
	env := os.Getenv(keyUpdateEnv)
	if env == "" {
		keyUpdateInterval = protocol.KeyUpdateInterval
		return
	}
	interval, err := strconv.ParseUint(env, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse %s: %s", keyUpdateEnv, err))
	}
	keyUpdateInterval = interval
}

type updatableAEAD struct {
	suite *qtls.CipherSuiteTLS13

	keyPhase          protocol.KeyPhase
	largestAcked      protocol.PacketNumber
	keyUpdateInterval uint64

	// Time when the keys should be dropped. Keys are dropped on the next call to Open().
	prevRcvAEADExpiry time.Time
	prevRcvAEAD       cipher.AEAD

	firstRcvdWithCurrentKey protocol.PacketNumber
	firstSentWithCurrentKey protocol.PacketNumber
	numRcvdWithCurrentKey   uint64
	numSentWithCurrentKey   uint64
	rcvAEAD                 cipher.AEAD
	sendAEAD                cipher.AEAD
	// caches cipher.AEAD.Overhead(). This speeds up calls to Overhead().
	aeadOverhead int

	nextRcvAEAD           cipher.AEAD
	nextSendAEAD          cipher.AEAD
	nextRcvTrafficSecret  []byte
	nextSendTrafficSecret []byte

	headerDecrypter headerProtector
	headerEncrypter headerProtector

	rttStats *congestion.RTTStats

	logger utils.Logger

	// use a single slice to avoid allocations
	nonceBuf []byte
}

var _ ShortHeaderOpener = &updatableAEAD{}
var _ ShortHeaderSealer = &updatableAEAD{}

func newUpdatableAEAD(rttStats *congestion.RTTStats, logger utils.Logger) *updatableAEAD {
	return &updatableAEAD{
		largestAcked:            protocol.InvalidPacketNumber,
		firstRcvdWithCurrentKey: protocol.InvalidPacketNumber,
		firstSentWithCurrentKey: protocol.InvalidPacketNumber,
		keyUpdateInterval:       keyUpdateInterval,
		rttStats:                rttStats,
		logger:                  logger,
	}
}

func (a *updatableAEAD) rollKeys(now time.Time) {
	a.keyPhase++
	a.firstRcvdWithCurrentKey = protocol.InvalidPacketNumber
	a.firstSentWithCurrentKey = protocol.InvalidPacketNumber
	a.numRcvdWithCurrentKey = 0
	a.numSentWithCurrentKey = 0
	a.prevRcvAEAD = a.rcvAEAD
	a.prevRcvAEADExpiry = now.Add(3 * a.rttStats.PTO())
	a.rcvAEAD = a.nextRcvAEAD
	a.sendAEAD = a.nextSendAEAD

	a.nextRcvTrafficSecret = a.getNextTrafficSecret(a.suite.Hash, a.nextRcvTrafficSecret)
	a.nextSendTrafficSecret = a.getNextTrafficSecret(a.suite.Hash, a.nextSendTrafficSecret)
	a.nextRcvAEAD = createAEAD(a.suite, a.nextRcvTrafficSecret)
	a.nextSendAEAD = createAEAD(a.suite, a.nextSendTrafficSecret)
}

func (a *updatableAEAD) getNextTrafficSecret(hash crypto.Hash, ts []byte) []byte {
	return qtls.HkdfExpandLabel(hash, ts, []byte{}, "traffic upd", hash.Size())
}

// For the client, this function is called before SetWriteKey.
// For the server, this function is called after SetWriteKey.
func (a *updatableAEAD) SetReadKey(suite *qtls.CipherSuiteTLS13, trafficSecret []byte) {
	a.rcvAEAD = createAEAD(suite, trafficSecret)
	a.headerDecrypter = newHeaderProtector(suite, trafficSecret, false)
	if a.suite == nil {
		a.nonceBuf = make([]byte, a.rcvAEAD.NonceSize())
		a.aeadOverhead = a.rcvAEAD.Overhead()
		a.suite = suite
	}

	a.nextRcvTrafficSecret = a.getNextTrafficSecret(suite.Hash, trafficSecret)
	a.nextRcvAEAD = createAEAD(suite, a.nextRcvTrafficSecret)
}

// For the client, this function is called after SetReadKey.
// For the server, this function is called before SetWriteKey.
func (a *updatableAEAD) SetWriteKey(suite *qtls.CipherSuiteTLS13, trafficSecret []byte) {
	a.sendAEAD = createAEAD(suite, trafficSecret)
	a.headerEncrypter = newHeaderProtector(suite, trafficSecret, false)
	if a.suite == nil {
		a.nonceBuf = make([]byte, a.sendAEAD.NonceSize())
		a.aeadOverhead = a.sendAEAD.Overhead()
		a.suite = suite
	}

	a.nextSendTrafficSecret = a.getNextTrafficSecret(suite.Hash, trafficSecret)
	a.nextSendAEAD = createAEAD(suite, a.nextSendTrafficSecret)
}

func (a *updatableAEAD) Open(dst, src []byte, rcvTime time.Time, pn protocol.PacketNumber, kp protocol.KeyPhaseBit, ad []byte) ([]byte, error) {
	if a.prevRcvAEAD != nil && rcvTime.After(a.prevRcvAEADExpiry) {
		a.prevRcvAEAD = nil
		a.prevRcvAEADExpiry = time.Time{}
	}
	binary.BigEndian.PutUint64(a.nonceBuf[len(a.nonceBuf)-8:], uint64(pn))
	if kp != a.keyPhase.Bit() {
		if a.firstRcvdWithCurrentKey == protocol.InvalidPacketNumber || pn < a.firstRcvdWithCurrentKey {
			if a.keyPhase == 0 {
				// This can only occur when the first packet received has key phase 1.
				// This is an error, since the key phase starts at 0,
				// and peers are only allowed to update keys after the handshake is confirmed.
				return nil, qerr.Error(qerr.ProtocolViolation, "wrong initial keyphase")
			}
			if a.prevRcvAEAD == nil {
				return nil, ErrKeysDropped
			}
			// we updated the key, but the peer hasn't updated yet
			dec, err := a.prevRcvAEAD.Open(dst, a.nonceBuf, src, ad)
			if err != nil {
				err = ErrDecryptionFailed
			}
			return dec, err
		}
		// try opening the packet with the next key phase
		dec, err := a.nextRcvAEAD.Open(dst, a.nonceBuf, src, ad)
		if err != nil {
			return nil, ErrDecryptionFailed
		}
		// Opening succeeded. Check if the peer was allowed to update.
		if a.firstSentWithCurrentKey == protocol.InvalidPacketNumber {
			return nil, qerr.Error(qerr.ProtocolViolation, "keys updated too quickly")
		}
		a.rollKeys(rcvTime)
		a.logger.Debugf("Peer updated keys to %s", a.keyPhase)
		a.firstRcvdWithCurrentKey = pn
		return dec, err
	}
	// The AEAD we're using here will be the qtls.aeadAESGCM13.
	// It uses the nonce provided here and XOR it with the IV.
	dec, err := a.rcvAEAD.Open(dst, a.nonceBuf, src, ad)
	if err != nil {
		err = ErrDecryptionFailed
	} else {
		a.numRcvdWithCurrentKey++
		if a.firstRcvdWithCurrentKey == protocol.InvalidPacketNumber {
			a.firstRcvdWithCurrentKey = pn
		}
	}
	return dec, err
}

func (a *updatableAEAD) Seal(dst, src []byte, pn protocol.PacketNumber, ad []byte) []byte {
	if a.firstSentWithCurrentKey == protocol.InvalidPacketNumber {
		a.firstSentWithCurrentKey = pn
	}
	a.numSentWithCurrentKey++
	binary.BigEndian.PutUint64(a.nonceBuf[len(a.nonceBuf)-8:], uint64(pn))
	// The AEAD we're using here will be the qtls.aeadAESGCM13.
	// It uses the nonce provided here and XOR it with the IV.
	return a.sendAEAD.Seal(dst, a.nonceBuf, src, ad)
}

func (a *updatableAEAD) SetLargestAcked(pn protocol.PacketNumber) {
	a.largestAcked = pn
}

func (a *updatableAEAD) updateAllowed() bool {
	return a.firstSentWithCurrentKey != protocol.InvalidPacketNumber &&
		a.largestAcked != protocol.InvalidPacketNumber &&
		a.largestAcked >= a.firstSentWithCurrentKey
}

func (a *updatableAEAD) shouldInitiateKeyUpdate() bool {
	if !a.updateAllowed() {
		return false
	}
	if a.numRcvdWithCurrentKey >= a.keyUpdateInterval {
		a.logger.Debugf("Received %d packets with current key phase. Initiating key update to the next key phase: %s", a.numRcvdWithCurrentKey, a.keyPhase+1)
		return true
	}
	if a.numSentWithCurrentKey >= a.keyUpdateInterval {
		a.logger.Debugf("Sent %d packets with current key phase. Initiating key update to the next key phase: %s", a.numSentWithCurrentKey, a.keyPhase+1)
		return true
	}
	return false
}

func (a *updatableAEAD) KeyPhase() protocol.KeyPhaseBit {
	if a.shouldInitiateKeyUpdate() {
		a.rollKeys(time.Now())
	}
	return a.keyPhase.Bit()
}

func (a *updatableAEAD) Overhead() int {
	return a.aeadOverhead
}

func (a *updatableAEAD) EncryptHeader(sample []byte, firstByte *byte, hdrBytes []byte) {
	a.headerEncrypter.EncryptHeader(sample, firstByte, hdrBytes)
}

func (a *updatableAEAD) DecryptHeader(sample []byte, firstByte *byte, hdrBytes []byte) {
	a.headerDecrypter.DecryptHeader(sample, firstByte, hdrBytes)
}
