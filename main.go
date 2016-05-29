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
		log.Printf("Error", err)
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
		// fmt.Printf("%s error: %s\n", url, err)
		return false, -1, err
	}

	headResponse.Body.Close()

	if isBadStatus(headResponse) {
		// fmt.Fprintf(os.Stderr, "Trying a GET request to retrieve %s\n", url)
		getResponse, err := client.Get(url)

		if err != nil {
			getResponse.Body.Close()
			// fmt.Printf("%s error: %s\n", url, err)
			return false, -1, err
		}

		getResponse.Body.Close()

		if isBadStatus(getResponse) {
			// fmt.Printf("%s %d\n", url, getResponse.StatusCode)
			return false, getResponse.StatusCode, err
		}
	}

	return true, headResponse.StatusCode, nil
}

func worker(id int, checkJobs <-chan Bookmark, failures chan<- LookupFailure, workgroup *sync.WaitGroup) {
	defer workgroup.Done()

	for bookmark := range checkJobs {
		fmt.Fprintf(os.Stdout, "Worker %02d: Processing job for url %s\n", id, bookmark.Href)
		valid, code, err := check(bookmark)
		if !valid {
			fmt.Fprintf(os.Stdout, "Worker %02d: ERROR: %s %d %s\n", id, bookmark.Href, code, err)
			failures <- LookupFailure{bookmark, err}
		} else {
			fmt.Fprintf(os.Stdout, "Worker %02d: Success for %s\n", id, bookmark.Href)
		}
	}
}

func failureReader(failures <-chan LookupFailure) {
	file, err := os.Create("failedlinks.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for failure := range failures {
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

func checkAll(bookmarkJson []byte) {
	// create job and results channel
	checkOperations := make(chan Bookmark, 10)
	failures := make(chan LookupFailure, 10)
	workgroup := new(sync.WaitGroup)

	// start failure reader
	go failureReader(failures)

	// start workers
	for w := 1; w <= 10; w++ {
		workgroup.Add(1)
		go worker(w, checkOperations, failures, workgroup)
	}

	// send off URLs to check
	for _, bookmark := range parseJson(bookmarkJson) {
		checkOperations <- bookmark
	}

	close(checkOperations)

	log.Println("No more check jobs written")

	workgroup.Wait()

	log.Println("Closing failure channel")

	close(failures)
}

func deleteBookmark(token string, bookmark Bookmark) {
	endpoint := buildDeleteEndpoint(token, bookmark.Href)

	log.Printf("Deleting %s\n", bookmark.Href)

	response, err := http.Get(endpoint)
	defer response.Body.Close()

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	fmt.Printf("%s", body)
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
	fmt.Printf("%s", string(bookmarks))
}

func handleDeleteAction(token string, resultsFileName string) {
	if resultsFileName == "-" {
		fmt.Println("Using stdin")
		deleteAll(token, os.Stdin)
	} else {
		fmt.Printf("Using bookmarks from %s\n", resultsFileName)
		file, err := os.Open(resultsFileName)
		if err != nil {
			log.Fatal("Could not read file with bookmarks to delete")
		} else {
			deleteAll(token, file)
		}
	}
}

func handleCheckAction(token string, inputBookmarks string) {
	var bookmarkJson []byte
	if len(inputBookmarks) > 0 {
		bookmarkJson, _ = ioutil.ReadFile(inputBookmarks)
	} else {
		bookmarkJson, _ = downloadBookmarks(token)
	}
	checkAll(bookmarkJson)
}

func main() {
	var downloadAction bool
	flag.BoolVar(&downloadAction, "download", false, "Download all bookmarks, write them to stdout")

	var deleteAction bool
	flag.BoolVar(&deleteAction, "delete", false, "Use this to delete bookmarks. Requires passing a list of links to delete.")

	var token string
	flag.StringVar(&token, "token", "", "Mandatory authentication token")

	var resultsSource string
	flag.StringVar(&resultsSource, "result", "-", "File to store results in, defaults to stdout")

	var inputBookmarks string
	flag.StringVar(&inputBookmarks, "inputBookmarks", "", "Path to exported bookmarks in JSON format")

	var checkAction bool
	flag.BoolVar(&checkAction, "check", false, "Check the links of all bookmarks")

	flag.Parse()

	// TODO: Validate that token got set
	if len(token) == 0 {
		log.Fatal("-token parameter has to be set")
	}

	if downloadAction {
		handleDownloadAction(token)
	}

	if deleteAction {
		handleDeleteAction(token, resultsSource)
	}

	if checkAction {
		handleCheckAction(token, inputBookmarks)
	}
}
