package main

import (
	"ManyACG-Bot/cmd"

	flag "github.com/spf13/pflag"
)

func main() {
	var update bool
	var version bool
	flag.BoolVar(&update, "update", false, "Update the program")
	flag.BoolVar(&version, "version", false, "Show the version")
	flag.Parse()
	if update {
		cmd.Update()
		return
	}
	if version {
		cmd.ShowVersion()
		return
	}
	cmd.Run()
}
