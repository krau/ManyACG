package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "github.com/krau/ManyACG",
	Short: "github.com/krau/ManyACG",
	Long:  "A Telegram bot for ACG channel.",
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
