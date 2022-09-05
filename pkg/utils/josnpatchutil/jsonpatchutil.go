package josnpatchutil

import (
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/mitchellh/mapstructure"
)

func Parse(raw []byte) (jsonpatch.Patch, error) {
	return jsonpatch.DecodePatch(raw)
}

func GetValue(patch jsonpatch.Operation, value interface{}) error {
	valueInterface, err := patch.ValueInterface()
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(valueInterface, value); err != nil {
		return err
	}
	return nil
}
