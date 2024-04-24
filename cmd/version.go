package cmd

import (
	. "ManyACG-Bot/logger"
)

func ShowVersion() {
	Logger.Infof("Version: %s", Version)
}
