package cmd

import (
	"fmt"
	"runtime"

	"github.com/krau/ManyACG/internal/common/version"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number of ManyACG",
	Run: func(cmd *cobra.Command, args []string) {
		ShowVersion()
	},
}

func init() {
	rootCmd.AddCommand(VersionCmd)
}

func ShowVersion() {
	fmt.Printf("ManyACG version: %s %s/%s\nBuildTime: %s, Commit: %s\n",
		version.Version, runtime.GOOS, runtime.GOARCH, version.BuildTime, version.Commit)
}
