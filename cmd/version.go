package cmd

import "fmt"

const (
	Version string = "0.2.0"
)

func ShowVersion() {
	fmt.Println(Version)
}
