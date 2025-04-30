/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package josnpatchutil

import (
	jsonpatch "github.com/evanphx/json-patch/v5"
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
