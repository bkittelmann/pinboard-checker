package cmd

import (
	"crypto/tls"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/bkittelmann/pinboard-checker/pinboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var inputFile string
var outputFile string

func init() {
	checkCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "File containing links to check. To read stdin use '-'.")
	checkCmd.Flags().String("inputFormat", "json", "Format of file with links. Can be either 'json' (default) or 'txt'")
	checkCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "-", "Where the report should be written to")
	checkCmd.Flags().String("outputFormat", "txt", "Allowed values are 'txt' (default) or 'json'")
	checkCmd.Flags().BoolP("verbose", "v", false, "Verbose logging, will report successful link lookups")
	checkCmd.Flags().Bool("noColor", false, "Do not use colorized status output")
	checkCmd.Flags().String("timeout", pinboard.DefaultTimeout.String(), "Timeout for HTTP client calls")
	checkCmd.Flags().Int("requestRate", pinboard.DefaultRequestRate, "How many HTTP requests are allowed simultaneously")
	checkCmd.Flags().Int("numberOfWorkers", pinboard.DefaultNumberOfWorkers, "How many concurrent workers are used")
	checkCmd.Flags().Bool("skipVerify", false, "If set, do not verify hosts of HTTPs domains. Avoids certificate errors in certain cases.")

	viper.BindPFlag("inputFormat", checkCmd.Flags().Lookup("inputFormat"))
	viper.BindPFlag("outputFormat", checkCmd.Flags().Lookup("outputFormat"))
	viper.BindPFlag("verbose", checkCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("noColor", checkCmd.Flags().Lookup("noColor"))
	viper.BindPFlag("timeout", checkCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("requestRate", checkCmd.Flags().Lookup("requestRate"))
	viper.BindPFlag("numberOfWorkers", checkCmd.Flags().Lookup("numberOfWorkers"))
	viper.BindPFlag("skipVerify", checkCmd.Flags().Lookup("skipVerify"))

	RootCmd.AddCommand(checkCmd)
}

func makeReporter(format pinboard.Format) pinboard.Reporter {
	verbose := viper.GetBool("verbose")
	noColor := viper.GetBool("noColor")
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
		inputFormatRaw := viper.GetString("inputFormat")
		inputFormat, inputErr := pinboard.FormatFromString(inputFormatRaw)
		if inputErr != nil {
			log.Fatalf("Invalid input format: %s", inputFormatRaw)
		}

		outputFormatRaw := viper.GetString("outputFormat")
		outputFormat, outputErr := pinboard.FormatFromString(outputFormatRaw)
		if outputErr != nil {
			log.Fatalf("Invalid output format: %s", outputFormatRaw)
		}

		// validate the timeout flag
		timeoutRaw := viper.GetString("timeout")
		timeout, parseErr := time.ParseDuration(timeoutRaw)
		if parseErr != nil {
			log.Fatalf("Invalid timeout value: %s", timeoutRaw)
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
			token := validateToken()
			endpoint := viper.GetString("endpoint")
			endpointUrl, _ := url.Parse(endpoint)

			client := pinboard.NewClient(token, endpointUrl)
			bookmarks, _ = client.GetAllBookmarks()
		}

		var tlsConfig *tls.Config
		if viper.GetBool("skipVerify") {
			tlsConfig = pinboard.TlsConfigAllowingInsecure()
		} else {
			tlsConfig = &tls.Config{}
		}

		checker := &pinboard.Checker{
			Reporter:        reporter,
			RequestRate:     viper.GetInt("requestRate"),
			NumberOfWorkers: viper.GetInt("numberOfWorkers"),

			Http: pinboard.DefaultHttpClient(timeout, tlsConfig),
		}
		checker.Run(bookmarks)
	},
}
