package cmd

import (
	"bufio"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.Flags().StringP("inputFile", "i", "-", "File containing links to delete")

	RootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Bulk-delete links stored in your pinboard",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		endpointUrl, _ := url.Parse(endpoint)
		inputFile, _ := cmd.Flags().GetString("inputFile")

		if inputFile == "-" {
			deleteAll(token, endpointUrl, os.Stdin)
		} else {
			file, err := os.Open(inputFile)
			if err != nil {
				log.Fatal("Could not read file with bookmarks to delete")
			} else {
				deleteAll(token, endpointUrl, file)
			}
		}
	},
}

func deleteAll(token string, endpoint *url.URL, reader io.Reader) {
	client := pinboard.NewClient(token, endpoint)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		client.DeleteBookmark(pinboard.Bookmark{Href: url, Description: ""})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
