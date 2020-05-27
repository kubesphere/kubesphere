package jwk

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// KeyUsageType is used to denote what this key should be used for
type KeyUsageType string

const (
	// ForSignature is the value used in the headers to indicate that
	// this key should be used for signatures
	ForSignature KeyUsageType = "sig"
	// ForEncryption is the value used in the headers to indicate that
	// this key should be used for encryptiong
	ForEncryption KeyUsageType = "enc"
)

// KeyOperation is used to denote the allowed operations for a Key
type KeyOperation string

// KeyOperationList represents an slice of KeyOperation
type KeyOperationList []KeyOperation

var keyOps = map[string]struct{}{"sign": {}, "verify": {}, "encrypt": {}, "decrypt": {}, "wrapKey": {}, "unwrapKey": {}, "deriveKey": {}, "deriveBits": {}}

// KeyOperation constants
const (
	KeyOpSign       KeyOperation = "sign"       // (compute digital signature or MAC)
	KeyOpVerify                  = "verify"     // (verify digital signature or MAC)
	KeyOpEncrypt                 = "encrypt"    // (encrypt content)
	KeyOpDecrypt                 = "decrypt"    // (decrypt content and validate decryption, if applicable)
	KeyOpWrapKey                 = "wrapKey"    // (encrypt key)
	KeyOpUnwrapKey               = "unwrapKey"  // (decrypt key and validate decryption, if applicable)
	KeyOpDeriveKey               = "deriveKey"  // (derive key)
	KeyOpDeriveBits              = "deriveBits" // (derive bits not to be used as a key)
)

// Accept determines if Key Operation is valid
func (keyOperationList *KeyOperationList) Accept(v interface{}) error {
	switch x := v.(type) {
	case KeyOperationList:
		*keyOperationList = x
		return nil
	default:
		return errors.Errorf(`invalid value %T`, v)
	}
}

// UnmarshalJSON unmarshals and checks data as KeyType Algorithm
func (keyOperationList *KeyOperationList) UnmarshalJSON(data []byte) error {
	var tempKeyOperationList []string
	err := json.Unmarshal(data, &tempKeyOperationList)
	if err != nil {
		return fmt.Errorf("invalid key operation")
	}
	for _, value := range tempKeyOperationList {
		_, ok := keyOps[value]
		if !ok {
			return fmt.Errorf("unknown key operation")
		}
		*keyOperationList = append(*keyOperationList, KeyOperation(value))
	}
	return nil
}
