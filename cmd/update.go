package cmd

import (
	"log"

	"github.com/krau/ManyACG/common"

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
	v := semver.MustParse(common.Version)
	latest, err := selfupdate.UpdateSelf(v, "krau/ManyACG")
	if err != nil {
		log.Println("Binary update failed:", err)
		return
	}
	if latest.Version.Equals(v) {
		log.Println("Current binary is the latest version", common.Version)
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
	}
}
