package cmd

import (
	"fmt"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "pinboard-checker",
	Short: "Tool for checking the state of your links on pinboard.in",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("token", "t", "", "The pinboard API token")
	RootCmd.PersistentFlags().String("endpoint", pinboard.DefaultEndpoint.String(), "URL of pinboard API endpoint")

	//  TODO: Use the viper config initialization
	//	cobra.OnInitialize(initConfig)
}
