package pinboard

import (
	"crypto/tls"
	"io"
	"io/ioutil"
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

func TlsConfigAllowingInsecure() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

func DefaultHttpClient(timeout time.Duration, tls *tls.Config) *http.Client {
	cookieJar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		TLSClientConfig: tls,
	}

	return &http.Client{
		Jar:       cookieJar,
		Timeout:   timeout,
		Transport: transport,
	}
}

func (checker *Checker) check(bookmark Bookmark) (bool, int, error) {
	url := bookmark.Href

	headRequest, _ := http.NewRequest(http.MethodHead, url, nil)
	headResponse, err := checker.Http.Do(headRequest)
	if err != nil {
		return false, -1, err
	}
	io.Copy(ioutil.Discard, headResponse.Body)
	headResponse.Body.Close()

	if isBadStatus(headResponse) {
		getRequest, _ := http.NewRequest(http.MethodGet, url, nil)
		getResponse, err := checker.Http.Do(getRequest)

		if err != nil {
			return false, -1, err
		}

		io.Copy(ioutil.Discard, getResponse.Body)
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
