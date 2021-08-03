package cmd

import (
	"github.com/spf13/cobra"

	"archive-my/api"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serve the backend API",
	Run:   serverFn,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func serverFn(cmd *cobra.Command, args []string) {
	a := api.App{}
	a.InitDeps()
	a.Run()
}
