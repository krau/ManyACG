package main

import (
	"ManyACG-Bot/cmd"

	flag "github.com/spf13/pflag"
)

func main() {
	var update bool
	flag.BoolVar(&update, "update", false, "Update the program")
	flag.Parse()
	if update {
		cmd.Update()
		return
	}
	cmd.Run()
}
