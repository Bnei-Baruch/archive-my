package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/subosito/gotenv"

	"github.com/Bnei-Baruch/archive-my/common"
)

var rootCmd = &cobra.Command{
	Use:   "archive-my",
	Short: "Personal data on archive",
	Long:  `Backend API for personal information of archive`,
}

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	gotenv.Load()
	common.Init()
}
