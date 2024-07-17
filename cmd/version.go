package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	Version string = "0.11.2"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ManyACG",
	Run: func(cmd *cobra.Command, args []string) {
		ShowVersion()
	},
}

func init() {
	rootCmd.AddCommand(VersionCmd)
}

func ShowVersion() {
	fmt.Printf("ManyACG version %s %s/%s", Version, runtime.GOOS, runtime.GOARCH)
}
