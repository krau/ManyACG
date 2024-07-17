package cmd

import (
	"log"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"up", "upgrade"},
	Short:   "Upgrade ManyACG",
	Run: func(cmd *cobra.Command, args []string) {
		Update()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func Update() {
	v := semver.MustParse(Version)
	latest, err := selfupdate.UpdateSelf(v, "krau/ManyACG")
	if err != nil {
		log.Println("Binary update failed:", err)
		return
	}
	if latest.Version.Equals(v) {
		log.Println("Current binary is the latest version", Version)
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
	}
}
