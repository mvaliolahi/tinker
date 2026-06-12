package api

import (
	"encoding/json"
	"fmt"
)

// decodeJSON deserializes JSON data into the provided value.
func decodeJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return json.Unmarshal(data, v)
}

// encodeJSON serializes the provided value to pretty-printed JSON.
func encodeJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
