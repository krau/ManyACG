//go:build !sonic && !stdjson

package objectuuid

import "github.com/goccy/go-json"

func MarshalJSON(val any) ([]byte, error) {
	return json.Marshal(val)
}

func UnmarshalJSON(data []byte, val any) error {
	return json.Unmarshal(data, val)
}
