package cmd

import (
	"io"
	"log"
	"net/url"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

var inputFile string
var inputFormatRaw string
var outputFile string
var outputFormatRaw string
var verbose bool
var noColor bool

func init() {
	checkCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "File containing links to check. To read stdin use '-'.")
	checkCmd.Flags().StringVar(&inputFormatRaw, "inputFormat", "json", "Format of file with links. Can be either 'json' (default) or 'txt'")
	checkCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "-", "Where the report should be written to")
	checkCmd.Flags().StringVar(&outputFormatRaw, "outputFormat", "txt", "Allowed values are 'txt' (default) or 'json'")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging, will report successful link lookups")
	checkCmd.Flags().BoolVar(&noColor, "noColor", false, "Do not use colorized status output")

	RootCmd.AddCommand(checkCmd)
}

func makeReporter(format pinboard.Format) pinboard.Reporter {
	var reporter pinboard.Reporter
	switch format {
	case pinboard.JSON:
		reporter = pinboard.NewJSONReporter(verbose)
	case pinboard.TXT:
		reporter = pinboard.NewSimpleFailureReporter(verbose, !noColor)
	}
	return reporter
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for stale links",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		// validate that format flags contain valid values
		inputFormat, inputErr := pinboard.FormatFromString(inputFormatRaw)
		if inputErr != nil {
			log.Fatalf("Invalid input format: %s", inputFormatRaw)
		}

		outputFormat, outputErr := pinboard.FormatFromString(outputFormatRaw)
		if outputErr != nil {
			log.Fatalf("Invalid output format: %s", outputFormatRaw)
		}

		reporter := makeReporter(outputFormat)

		var bookmarks []pinboard.Bookmark
		if len(inputFile) > 0 {
			var file io.Reader
			if inputFile == "-" {
				file = os.Stdin
			} else {
				file, _ = os.Open(inputFile)
			}
			bookmarks = pinboard.GetBookmarksFromFile(file, inputFormat)
		} else {
			token := validateToken(cmd)
			endpoint, _ := cmd.Flags().GetString("endpoint")
			endpointUrl, _ := url.Parse(endpoint)

			client := pinboard.NewClient(token, endpointUrl)
			bookmarks, _ = client.GetAllBookmarks()
		}

		pinboard.CheckAll(bookmarks, reporter)
	},
}
