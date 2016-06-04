package pinboardchecker

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
)

type LookupFailure struct {
	Bookmark Bookmark
	Error    error
}

type Reporter interface {
	onFailure(failure LookupFailure)
	onSuccess(bookmark Bookmark)
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

func worker(id int, checkJobs <-chan Bookmark, reporter Reporter, workgroup *sync.WaitGroup) {
	defer workgroup.Done()

	for bookmark := range checkJobs {
		debug("Worker %02d: Processing job for url %s", id, bookmark.Href)
		valid, code, err := check(bookmark)
		if !valid {
			reporter.onFailure(LookupFailure{bookmark, err})
			debug("Worker %02d: ERROR: %s %d %s", id, bookmark.Href, code, err)
		} else {
			reporter.onSuccess(bookmark)
			debug("Worker %02d: Success for %s\n", id, bookmark.Href)
		}
	}
}

func checkAll(bookmarks []Bookmark, reporter Reporter) {
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
