package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Bnei-Baruch/archive-my/api"
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
	a := new(api.App)
	a.Initialize()
	a.Run()
}
