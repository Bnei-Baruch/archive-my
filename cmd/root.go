package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"archive-my/pkg/utils"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "my-archive-api",
	Short: "Personal data on archive",
	Long:  `Backend API for personal information of archive`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(InitConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.toml)")
}

func InitConfig() {
	if err := utils.InitConfig(cfgFile, ""); err != nil {
		panic(errors.Wrapf(err, "Could not read config, using: %s", viper.ConfigFileUsed()))
	}
}
