package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Bnei-Baruch/archive-my/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of archive-my",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Archive personalization service version %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
