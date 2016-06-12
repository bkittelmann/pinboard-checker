package cmd

import (
	"os"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
)

var token string
var inputFile string
var outputFile string
var verbose bool
var noColor bool

func init() {
	checkCmd.Flags().StringVarP(&token, "token", "t", "", "The pinboard API token")
	checkCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "File containing links to check")
	checkCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "-", "Where the report should be written to")
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging, will report successful link lookups")
	checkCmd.Flags().BoolVar(&noColor, "noColor", false, "Do not use colorized status output")

	RootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for stale links",
	Long:  "...",

	Run: func(cmd *cobra.Command, args []string) {
		var bookmarks []pinboard.Bookmark
		if len(inputFile) > 0 {
			bookmarkJson, _ := os.Open(inputFile)
			bookmarks = pinboard.ParseJSON(bookmarkJson)
		} else {
			bookmarks, _ = pinboard.GetAllBookmarks(token)
		}

		// different failure reporter depending on setting of outputFile, default to
		// stderr simple error printing for now
		var reporter pinboard.Reporter
		switch {
		default:
			reporter = pinboard.NewSimpleFailureReporter(verbose, !noColor)
		}

		pinboard.CheckAll(bookmarks, reporter)
	},
}
