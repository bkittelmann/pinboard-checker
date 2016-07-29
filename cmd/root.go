package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger = logrus.New()

type OutputFormatter struct{}

func (o *OutputFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())
	return []byte(fmt.Sprintf("%s: %s\n", level, entry.Message)), nil
}

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
	// configure logger
	logger.Formatter = &OutputFormatter{}

	// configure flags
	RootCmd.PersistentFlags().StringP("token", "t", "", "The pinboard API token")
	RootCmd.PersistentFlags().String("endpoint", pinboard.DefaultEndpoint.String(), "URL of pinboard API endpoint")

	// initialize Viper to set flags from content in config files
	viper.SetConfigName("pinboard-checker")
	viper.AddConfigPath("./")
	viper.AddConfigPath("$HOME/.pinboard-checker")

	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.UnsupportedConfigError:
			// do nothing, this means no configuration is available, but flags could still be set
		default:
			logger.Fatalf("Error reading config file %s: %s\n", viper.ConfigFileUsed(), err.Error())
		}
	}

	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("endpoint", RootCmd.PersistentFlags().Lookup("endpoint"))

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PINBOARD_CHECKER")
}

func validateToken() string {
	token := viper.GetString("token")
	if len(token) == 0 {
		logger.Fatal("Token flag is mandatory for export command")
	}
	return token
}
