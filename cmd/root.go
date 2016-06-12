package cmd

import (
	"fmt"
	"os"

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
	//  TODO: Use the viper config initialization
	//	cobra.OnInitialize(initConfig)
}
