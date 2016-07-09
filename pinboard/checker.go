package pinboard

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/juju/ratelimit"
)

type LookupFailure struct {
	Bookmark Bookmark
	Code     int
	Error    error
}

type Reporter interface {
	onFailure(failure LookupFailure)
	onSuccess(bookmark Bookmark)
	onEnd()
}

var CheckTimeout = 10 * time.Second
var RequestsPerSecond float64 = 10
var DefaultNumberOfWorkers = 10

// we consider HTTP 429 indicative that the resource exists
func isBadStatus(response *http.Response) bool {
	return response.StatusCode != 200 && response.StatusCode != http.StatusTooManyRequests
}

func check(bookmark Bookmark, timeout time.Duration) (bool, int, error) {
	cookieJar, _ := cookiejar.New(nil)

	// TODO: Use same client in all workers
	client := &http.Client{
		Jar:     cookieJar,
		Timeout: timeout,
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
			return false, -1, err
		}

		getResponse.Body.Close()

		if isBadStatus(getResponse) {
			return false, getResponse.StatusCode, err
		}
	}

	return true, headResponse.StatusCode, nil
}

func worker(id int, checkJobs <-chan Bookmark, reporter Reporter, workgroup *sync.WaitGroup, timeout time.Duration, tokenBucket *ratelimit.Bucket) {
	defer workgroup.Done()

	for bookmark := range checkJobs {
		tokenBucket.Wait(1)
		debug("Worker %02d: Processing job for url %s", id, bookmark.Href)
		valid, code, err := check(bookmark, timeout)
		if !valid {
			reporter.onFailure(LookupFailure{bookmark, code, err})
			debug("Worker %02d: ERROR: %s %d %s", id, bookmark.Href, code, err)
		} else {
			reporter.onSuccess(bookmark)
			debug("Worker %02d: Success for %s\n", id, bookmark.Href)
		}
	}
}

func CheckAll(bookmarks []Bookmark, reporter Reporter, timeout time.Duration, requestRate float64, numberOfWorkers int) {
	jobs := make(chan Bookmark, numberOfWorkers)
	workgroup := new(sync.WaitGroup)
	tokenBucket := ratelimit.NewBucketWithRate(requestRate, int64(requestRate))

	// start workers
	for w := 1; w <= numberOfWorkers; w++ {
		workgroup.Add(1)
		go worker(w, jobs, reporter, workgroup, timeout, tokenBucket)
	}

	// send off URLs to check
	for _, bookmark := range bookmarks {
		jobs <- bookmark
	}

	close(jobs)
	workgroup.Wait()
	reporter.onEnd()
}
