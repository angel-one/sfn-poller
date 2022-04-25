package utils

import (
	"encoding/json"
)

func Marshal(v interface{}) (*string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

func Unmarshal(buf *string, v interface{}) error {
	err := json.Unmarshal([]byte(*buf), v)
	if err != nil {
		return err
	}
	return nil
}
