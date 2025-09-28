package log

import (
	"fmt"
	"time"
)

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func nowLocal(t time.Time) time.Time {
	return t.Local()
}
