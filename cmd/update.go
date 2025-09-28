package cmd

// import (
// 	"log"

// 	"strings"

// 	"github.com/krau/ManyACG/internal/common"

// 	"github.com/blang/semver"
// 	"github.com/rhysd/go-github-selfupdate/selfupdate"
// 	"github.com/spf13/cobra"
// )

// var updateCmd = &cobra.Command{
// 	Use:     "update",
// 	Aliases: []string{"up", "upgrade"},
// 	Short:   "Upgrade ManyACG",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		Update()
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(updateCmd)
// }

// func Update() {
// 	v := semver.MustParse(strings.TrimPrefix(common.Version, "v"))
// 	release, found, err := selfupdate.DetectLatest("krau/ManyACG")
// 	if err != nil {
// 		log.Println("Error occurred while detecting version:", err)
// 		return
// 	}
// 	if !found {
// 		log.Println("No newer version found")
// 		return
// 	}
// 	if release.Version.Equals(v) {
// 		log.Println("Current binary is the latest version", common.Version)
// 		return
// 	}
// 	if release.Version.LT(v) {
// 		log.Println("Current binary version", common.Version, "is newer than latest release", release.Version)
// 		return
// 	}
// 	// check major version
// 	if v.Major != release.Version.Major {
// 		log.Printf("New major version %s detected. Please check the release note and upgrade manually if necessary.\n", release.Version)
// 		return
// 	}
// 	latest, err := selfupdate.UpdateSelf(v, "krau/ManyACG")
// 	if err != nil {
// 		log.Println("Binary update failed:", err)
// 		return
// 	}
// 	if latest.Version.Equals(v) {
// 		log.Println("Current binary is the latest version", common.Version)
// 		return
// 	}
// 	log.Println("Successfully updated to version", latest.Version)
// 	log.Println("Release note:\n", latest.ReleaseNotes)
// }
