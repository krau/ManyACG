package cmd

import "fmt"

const (
	Version string = "v0.1.9"
)

func ShowVersion() {
	fmt.Println(Version)
}
