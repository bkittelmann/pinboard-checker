package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

func debug(format string, args ...interface{}) {
	if debugEnabled {
		log.Printf(format+"\n", args...)
	}
}

var debugEnabled bool

func readUrlsFromFile(source string) []string {
	urls := make([]string, 0)

	if file, err := os.Open(source); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			urls = append(urls, url)
		}

		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("ERROR: %s", err)
	}
	return urls
}

func deleteAll(token string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		deleteBookmark(token, Bookmark{Href: url, Description: ""})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func handleDownloadAction(token string) {
	readCloser, err := downloadBookmarks(token)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, readCloser)
	readCloser.Close()
}

func handleDeleteAction(token string, resultsFileName string) {
	if resultsFileName == "-" {
		debug("Using stdin")
		deleteAll(token, os.Stdin)
	} else {
		debug("Using bookmarks from %s\n", resultsFileName)
		file, err := os.Open(resultsFileName)
		if err != nil {
			log.Fatal("Could not read file with bookmarks to delete")
		} else {
			deleteAll(token, file)
		}
	}
}

func handleCheckAction(token string, inputFile string, outputFile string, verbose bool, noColor bool) {
	var bookmarks []Bookmark
	if len(inputFile) > 0 {
		bookmarkJson, _ := os.Open(inputFile)
		bookmarks = parseJson(bookmarkJson)
	} else {
		bookmarks, _ = getAllBookmarks(token)
	}

	// different failure reporter depending on setting of outputFile, default to
	// stderr simple error printing for now
	var reporter Reporter
	switch {
	default:
		reporter = newSimpleFailureReporter(verbose, !noColor)
	}

	checkAll(bookmarks, reporter)
}

func main() {
	var downloadAction bool
	flag.BoolVar(&downloadAction, "download", false, "Download all bookmarks, write them to stdout")

	var deleteAction bool
	flag.BoolVar(&deleteAction, "delete", false, "Use this to delete bookmarks. Requires passing a list of links to delete.")

	var token string
	flag.StringVar(&token, "token", "", "Mandatory authentication token")

	flag.BoolVar(&debugEnabled, "debug", false, "Enable debug logs, will be printed on stderr")

	var outputFile string
	flag.StringVar(&outputFile, "outputFile", "-", "File to store results of check operation in, defaults to stdout")

	var inputFile string
	flag.StringVar(&inputFile, "inputFile", "", "File containing bookmarks to check. If empty it will download all bookmarks from pinboard.")

	var inputFormat string
	flag.StringVar(&inputFormat, "inputFormat", "text", "Which format the input file is in (can be 'text', 'json')")

	var checkAction bool
	flag.BoolVar(&checkAction, "check", false, "Check the links of all bookmarks")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Report successful link lookups")

	var noColor bool
	flag.BoolVar(&noColor, "no-color", false, "Disable colorized output for text output")

	flag.Parse()

	// at least one action flag needs to be set, print usage if no flags are present
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if len(token) == 0 {
		log.Fatal("-token parameter has to be set")
	}

	if downloadAction {
		handleDownloadAction(token)
	}

	if deleteAction {
		handleDeleteAction(token, outputFile)
	}

	if checkAction {
		handleCheckAction(token, inputFile, outputFile, verbose, noColor)
	}
}
