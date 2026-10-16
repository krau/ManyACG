package cmd

import (
	"fmt"
	"os"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/spf13/cobra"
)

const mainCommandAnnotation = "manyacg.requires-main"

var rootCmd = &cobra.Command{
	Use:   "manyacg",
	Short: "manyacg",
	// Failures such as a forbidden role are not usage errors, so cobra should not print the help text.
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		for current := cmd; current != nil; current = current.Parent() {
			if current.Annotations[mainCommandAnnotation] != "true" {
				continue
			}
			if !runtimecfg.Get().App.IsMain() {
				return fmt.Errorf("%s requires app.role = main", cmd.CommandPath())
			}
			break
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func init() {
	// Subcommands defining their own PersistentPreRunE would otherwise skip the role check above.
	cobra.EnableTraverseRunHooks = true
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
