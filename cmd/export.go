package cmd

import (
	"io"
	"log"
	"net/url"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Download your bookmarks",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		token := validateToken()

		endpoint := viper.GetString("endpoint")
		endpointUrl, _ := url.Parse(endpoint)

		client := pinboard.NewClient(token, endpointUrl)

		readCloser, err := client.DownloadBookmarks()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, readCloser)
		readCloser.Close()
	},
}
