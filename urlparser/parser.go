package urlparser

import (
	"encoding/json"
	"os"
)

func ReadFile(filename string) ([]byte, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func ParseJSON(js []byte) ([]string, error) {
	var s []string
	if err := json.Unmarshal(js, &s); err != nil {
		return nil, err
	}

	return s, nil
}
