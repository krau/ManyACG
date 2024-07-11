package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version string = "0.10.10"
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
	fmt.Println(Version)
}
