package cmd

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.Flags().StringP("token", "t", "", "The pinboard API token")
	exportCmd.Flags().String("endpoint", pinboard.DefaultEndpoint.String(), "URL of pinboard API endpoint")

	RootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Download your bookmarks",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		if len(token) == 0 {
			fmt.Println("ERROR: Token flag is mandatory for export command")
			os.Exit(1)
		}

		endpoint, _ := cmd.Flags().GetString("endpoint")
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
