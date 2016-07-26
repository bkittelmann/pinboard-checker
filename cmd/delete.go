package cmd

import (
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	deleteCmd.Flags().StringP("inputFile", "i", "", "File containing URLs to delete.")

	RootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete [urls ...]",
	Short: "Bulk-delete links stored in your pinboard",
	Long: `Delete bookmarks in your pinboard.

Simply provide the URLs of the bookmarks to delete as arguments to
this command. Alternatively, you can use a file that stores the URLs
as input (see -i flag). 

If you use the -i flag, the file must have one URL per line. 
To read from stdout, use '-' as file name.`,

	Run: func(cmd *cobra.Command, args []string) {
		token := validateToken()
		endpoint := viper.GetString("endpoint")
		endpointUrl, _ := url.Parse(endpoint)

		var reader io.Reader

		if len(args) > 0 {
			allArgs := strings.Join(args, "\n")
			reader = strings.NewReader(allArgs)
		} else {
			inputFile, _ := cmd.Flags().GetString("inputFile")
			switch inputFile {
			case "":
				log.Fatalf("Invalid: No args given, and no inputFile parameter used")
			case "-":
				reader = os.Stdin
			default:
				file, err := os.Open(inputFile)
				if err != nil {
					log.Fatal("Could not read file with bookmarks to delete")
				} else {
					reader = file
				}
			}
		}

		deleteErr := deleteAll(token, endpointUrl, pinboard.ParseText(reader))
		if deleteErr != nil {
			os.Exit(1)
		}
	},
}

func deleteAll(token string, endpoint *url.URL, bookmarks []pinboard.Bookmark) (err error) {
	client := pinboard.NewClient(token, endpoint)
	var errorDuringDelete bool = false
	for _, bookmark := range bookmarks {
		delErr := client.DeleteBookmark(bookmark)
		if delErr != nil {
			log.Printf("Error trying to delete %s: %s", bookmark.Href, delErr)
			errorDuringDelete = true
		}
	}
	if errorDuringDelete {
		err = errors.New("Encountered at least one error when trying to delete bookmarks")
	}
	return
}
