package cmd

import (
	"log"
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

var token string
var inputFile string
var outputFile string
var outputFormat string
var verbose bool
var noColor bool

func init() {
	checkCmd.Flags().StringVarP(&token, "token", "t", "", "The pinboard API token")
	checkCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "File containing links to check")
	checkCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "-", "Where the report should be written to")
	checkCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "txt", "Allowed values are 'txt' (default) or 'json'")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging, will report successful link lookups")
	checkCmd.Flags().BoolVar(&noColor, "noColor", false, "Do not use colorized status output")

	RootCmd.AddCommand(checkCmd)
}

func makeReporter(format string) pinboard.Reporter {
	var reporter pinboard.Reporter
	switch format {
	case "json":
		reporter = pinboard.NewJSONReporter(verbose)
	case "txt":
		reporter = pinboard.NewSimpleFailureReporter(verbose, !noColor)
	default:
		log.Fatalf("'%s' is not a valid value for 'outputFormat'", format)
	}
	return reporter
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for stale links",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		reporter := makeReporter(outputFormat)

		var bookmarks []pinboard.Bookmark
		if len(inputFile) > 0 {
			bookmarkJson, _ := os.Open(inputFile)
			bookmarks = pinboard.ParseJSON(bookmarkJson)
		} else {
			bookmarks, _ = pinboard.GetAllBookmarks(token)
		}

		pinboard.CheckAll(bookmarks, reporter)
	},
}
