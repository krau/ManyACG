package cmd

import (
	"fmt"
	"os"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/spf13/cobra"
)

const writeCommandAnnotation = "manyacg.requires-write"

var rootCmd = &cobra.Command{
	Use:   "manyacg",
	Short: "manyacg",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		for current := cmd; current != nil; current = current.Parent() {
			if current.Annotations[writeCommandAnnotation] != "true" {
				continue
			}
			if !runtimecfg.Get().App.Capabilities.Has(runtimecfg.CapabilityWrite) {
				return fmt.Errorf("%s requires the write capability", cmd.CommandPath())
			}
			break
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
