package intutil

import "errors"

func Convert(subject interface{}) (int, error) {
	var result int

	switch subject.(type) {
	case uint64:
		result = int(subject.(uint64))
	case int64:
		result = int(subject.(int64))
	case int:
		result = subject.(int)
	default:
		return 0, errors.New("It is not a numeric input")
	}

	return result, nil
}
