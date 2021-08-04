package cmd

import (
	"github.com/spf13/cobra"

	"archive-my/api"
	"archive-my/pkg/chronicles"
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
	a.InitDeps()
	chr := new(chronicles.Chronicles)
	chr.Run()
	a.Run()
}
