package convert

import "encoding/json"

// ToMap attempts to decode JSON bytes into a generic map/slice structure.
func ToMap(data []byte) (any, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
