package cmd

import "fmt"

const (
	Version string = "0.2.1"
)

func ShowVersion() {
	fmt.Println(Version)
}
