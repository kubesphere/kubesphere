package protocol

// KeyPhase is the key phase
type KeyPhase uint64

func (p KeyPhase) Bit() KeyPhaseBit {
	return p%2 == 1
}

// KeyPhaseBit is the key phase bit
type KeyPhaseBit bool

const (
	// KeyPhaseZero is key phase 0
	KeyPhaseZero KeyPhaseBit = false
	// KeyPhaseOne is key phase 1
	KeyPhaseOne KeyPhaseBit = true
)

func (p KeyPhaseBit) String() string {
	if p == KeyPhaseZero {
		return "0"
	}
	return "1"
}
