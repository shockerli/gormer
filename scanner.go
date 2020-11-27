package gormer

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONScanner decoding DB field(JSON string) to custom type
func JSONScanner(f interface{}, value interface{}) error {
	switch value := value.(type) {
	case []byte:
		_ = json.Unmarshal(value, &f)
	case string:
		_ = json.Unmarshal([]byte(value), &f)
	}
	return nil
}

// JSONValuer encoding custom type to JSON string
func JSONValuer(f interface{}) (driver.Value, error) {
	if f == nil {
		return "{}", nil
	}
	s, err := json.Marshal(f)
	if err != nil {
		return "{}", err
	}

	return string(s), nil
}
