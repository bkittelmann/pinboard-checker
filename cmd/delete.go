package cmd

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.Flags().StringP("token", "t", "", "The pinboard API token")
	deleteCmd.Flags().StringP("inputFile", "i", "-", "File containing links to delete")

	RootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Bulk-delete links stored in your pinboard",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		inputFile, _ := cmd.Flags().GetString("inputFile")

		if inputFile == "-" {
			deleteAll(token, os.Stdin)
		} else {
			file, err := os.Open(inputFile)
			if err != nil {
				log.Fatal("Could not read file with bookmarks to delete")
			} else {
				deleteAll(token, file)
			}
		}
	},
}

func deleteAll(token string, reader io.Reader) {
	client := pinboard.NewClient(token, pinboard.DefaultEndpoint)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		client.DeleteBookmark(pinboard.Bookmark{Href: url, Description: ""})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
