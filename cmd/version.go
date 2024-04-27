package cmd

import "fmt"

const (
	Version string = "v0.1.8"
)

func ShowVersion() {
	fmt.Println(Version)
}
