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

var DefaultTimeout = 10 * time.Second
var DefaultRequestRate = 10
var DefaultNumberOfWorkers = 10

// we consider HTTP 429 indicative that the resource exists
func isBadStatus(response *http.Response) bool {
	return response.StatusCode != 200 && response.StatusCode != http.StatusTooManyRequests
}

type Checker struct {
	Reporter        Reporter
	RequestRate     int
	NumberOfWorkers int

	Http *http.Client
}

func DefaultHttpClient(timeout time.Duration) *http.Client {
	cookieJar, _ := cookiejar.New(nil)

	return &http.Client{
		Jar:     cookieJar,
		Timeout: timeout,
	}
}

func (checker *Checker) check(bookmark Bookmark) (bool, int, error) {
	url := bookmark.Href

	headResponse, err := checker.Http.Head(url)
	if err != nil {
		return false, -1, err
	}

	headResponse.Body.Close()

	if isBadStatus(headResponse) {
		getResponse, err := checker.Http.Get(url)

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

func (checker *Checker) worker(id int, checkJobs <-chan Bookmark, workgroup *sync.WaitGroup, tokenBucket *ratelimit.Bucket) {
	defer workgroup.Done()

	for bookmark := range checkJobs {
		tokenBucket.Wait(1)
		debug("Worker %02d: Processing job for url %s", id, bookmark.Href)
		valid, code, err := checker.check(bookmark)
		if !valid {
			checker.Reporter.onFailure(LookupFailure{bookmark, code, err})
			debug("Worker %02d: ERROR: %s %d %s", id, bookmark.Href, code, err)
		} else {
			checker.Reporter.onSuccess(bookmark)
			debug("Worker %02d: Success for %s\n", id, bookmark.Href)
		}
	}
}

func (checker *Checker) Run(bookmarks []Bookmark) {

	jobs := make(chan Bookmark, checker.NumberOfWorkers)
	workgroup := new(sync.WaitGroup)
	tokenBucket := ratelimit.NewBucketWithRate(float64(checker.RequestRate), int64(checker.RequestRate))

	// start workers
	for w := 1; w <= checker.NumberOfWorkers; w++ {
		workgroup.Add(1)
		go checker.worker(w, jobs, workgroup, tokenBucket)
	}

	// send off URLs to check
	for _, bookmark := range bookmarks {
		jobs <- bookmark
	}

	close(jobs)
	workgroup.Wait()
	checker.Reporter.onEnd()
}
