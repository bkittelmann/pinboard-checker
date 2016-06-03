package pinboardchecker

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
)

type FailureReporter func(LookupFailure)

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

func simpleFailureReporter(writers ...io.Writer) FailureReporter {
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	return func(failure LookupFailure) {
		for _, writer := range writers {
			fmt.Fprintf(writer, "[ERR] %s\n", failure.Bookmark.Href)
		}
	}
}

func checkAll(bookmarks []Bookmark, reporter FailureReporter) {
	jobs := make(chan Bookmark, 10)
	workgroup := new(sync.WaitGroup)

	// start workers
	for w := 1; w <= 10; w++ {
		workgroup.Add(1)
		go worker(w, jobs, reporter, workgroup)
	}

	// send off URLs to check
	for _, bookmark := range bookmarks {
		jobs <- bookmark
	}

	close(jobs)
	workgroup.Wait()
}
