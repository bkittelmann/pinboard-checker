package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Bookmark struct {
	Href        string
	Description string
}

type LookupFailure struct {
	Bookmark Bookmark
	Error    error
}

type FailureReporter func(LookupFailure)

func debug(format string, args ...interface{}) {
	if debugEnabled {
		log.Printf(format+"\n", args...)
	}
}

var debugEnabled bool

func parseJson(bookmarkJson []byte) []Bookmark {
	var bookmarks []Bookmark
	json.Unmarshal(bookmarkJson, &bookmarks)
	return bookmarks
}

func buildDownloadEndpoint(token string) string {
	endpoint, _ := url.Parse("https://api.pinboard.in/v1/posts/all")
	query := endpoint.Query()
	query.Add("auth_token", token)
	query.Add("format", "json")
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func buildDeleteEndpoint(token string, rawUrl string) string {
	endpoint, _ := url.Parse("https://api.pinboard.in/v1/posts/delete")
	query := endpoint.Query()
	query.Add("auth_token", token)
	query.Add("format", "json")
	query.Add("url", rawUrl)
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func downloadBookmarks(token string) ([]byte, error) {
	response, err := http.Get(buildDownloadEndpoint(token))
	defer response.Body.Close()

	if err != nil {
		debug("Error %s", err)
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

// we consider HTTP 429 indicative that the resource exists
func isBadStatus(response *http.Response) bool {
	return response.StatusCode != 200 && response.StatusCode != http.StatusTooManyRequests
}

func check(bookmark Bookmark) (bool, int, error) {
	cookieJar, _ := cookiejar.New(nil)

	// TODO: Use same client in all workers
	client := &http.Client{
		Jar: cookieJar,
	}

	url := bookmark.Href
	headResponse, err := client.Head(url)
	if err != nil {
		return false, -1, err
	}

	headResponse.Body.Close()

	if isBadStatus(headResponse) {
		getResponse, err := client.Get(url)

		if err != nil {
			getResponse.Body.Close()
			return false, -1, err
		}

		getResponse.Body.Close()

		if isBadStatus(getResponse) {
			return false, getResponse.StatusCode, err
		}
	}

	return true, headResponse.StatusCode, nil
}

func worker(id int, checkJobs <-chan Bookmark, reporter FailureReporter, workgroup *sync.WaitGroup) {
	defer workgroup.Done()

	for bookmark := range checkJobs {
		debug("Worker %02d: Processing job for url %s", id, bookmark.Href)
		valid, code, err := check(bookmark)
		if !valid {
			reporter(LookupFailure{bookmark, err})
			debug("Worker %02d: ERROR: %s %d %s", id, bookmark.Href, code, err)
		} else {
			debug("Worker %02d: Success for %s\n", id, bookmark.Href)
		}
	}
}

func csvFailureReader(failure LookupFailure) {
	file, err := os.Create("failedlinks.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var errorValue string
	if failure.Error != nil {
		errorValue = failure.Error.Error()
	}

	record := []string{
		failure.Bookmark.Description,
		failure.Bookmark.Href,
		errorValue,
	}
	writer.Write(record)
}

func stdoutFailureReporter(failure LookupFailure) {
	fmt.Fprintf(os.Stdout, "[ERR] %s\n", failure.Bookmark.Href)
}

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

func checkAll(bookmarkJson []byte, reporter FailureReporter) {
	jobs := make(chan Bookmark, 10)
	workgroup := new(sync.WaitGroup)

	// start workers
	for w := 1; w <= 10; w++ {
		workgroup.Add(1)
		go worker(w, jobs, reporter, workgroup)
	}

	// send off URLs to check
	for _, bookmark := range parseJson(bookmarkJson) {
		jobs <- bookmark
	}

	close(jobs)
	workgroup.Wait()
}

func deleteBookmark(token string, bookmark Bookmark) {
	endpoint := buildDeleteEndpoint(token, bookmark.Href)

	debug("Deleting %s\n", bookmark.Href)

	response, err := http.Get(endpoint)
	defer response.Body.Close()

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	debug("%s", body)
}

func deleteAll(token string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		deleteBookmark(token, Bookmark{url, ""})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func handleDownloadAction(token string) {
	bookmarks, err := downloadBookmarks(token)
	if err != nil {
		log.Fatal(err)
	}
	debug("%s", string(bookmarks))
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

func handleCheckAction(token string, inputFile string, outputFile string) {
	var bookmarkJson []byte
	if len(inputFile) > 0 {
		bookmarkJson, _ = ioutil.ReadFile(inputFile)
	} else {
		bookmarkJson, _ = downloadBookmarks(token)
	}

	// different failure reporter depending on setting of outputFile, default to
	// stderr simple error printing for now
	var reporter FailureReporter
	switch {
	default:
		reporter = stdoutFailureReporter
	}

	checkAll(bookmarkJson, reporter)
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
		handleCheckAction(token, inputFile, outputFile)
	}
}
