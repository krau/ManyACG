//go:build sonic

package objectuuid

import "github.com/bytedance/sonic"

func MarshalJSON(val any) ([]byte, error) {
	return sonic.Marshal(val)
}

func UnmarshalJSON(data []byte, val any) error {
	return sonic.Unmarshal(data, val)
}
