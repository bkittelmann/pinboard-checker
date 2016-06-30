package cmd

import (
	"io"
	"log"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.Flags().StringP("token", "t", "", "The pinboard API token")
	// add new API endpoint flag here, default it to default endpoint

	RootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Download your bookmarks",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		client := pinboard.NewClient(token, pinboard.DefaultEndpoint)
		readCloser, err := client.DownloadBookmarks()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, readCloser)
		readCloser.Close()
	},
}
