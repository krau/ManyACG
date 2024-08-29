package common

import (
	"time"
)

var (
	Version   string = "dev"
	BuildTime string = "Unknown"
	Commit    string = "Unknown"
)

func init() {
	if BuildTime == "Unknown" {
		BuildTime = time.Now().Format("2006-01-02 15:04:05")
	}
}
