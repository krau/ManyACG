//go:build stdjson

package objectuuid

import "encoding/json"

func MarshalJSON(val any) ([]byte, error) {
	return json.Marshal(val)
}

func UnmarshalJSON(data []byte, val any) error {
	return json.Unmarshal(data, val)
}
