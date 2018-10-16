package errors

import (
	"encoding/json"
	"errors"
)

func Wrap(data []byte) error {
	var j map[string]string
	err := json.Unmarshal(data, &j)
	if err != nil {
		return errors.New(string(data))
	} else if message := j["message"]; message != "" {
		return errors.New(message)
	} else if message := j["Error"]; message != "" {
		return errors.New(message)
	} else {
		return errors.New(string(data))
	}
}
