package pinboard

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func makeChecker() *Checker {
	return &Checker{
		RequestRate:     DefaultRequestRate,
		NumberOfWorkers: DefaultNumberOfWorkers,
		Http:            DefaultHttpClient(DefaultTimeout, TlsConfigAllowingInsecure()),
	}
}

// statusServer returns a handler that reads the desired status code from
// paths shaped like "/status/NNN" and treats "/" as 200 OK.
func statusServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/status/") {
			code, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/status/"))
			if err != nil {
				http.Error(w, "bad status path", http.StatusBadRequest)
				return
			}
			w.WriteHeader(code)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCheckGoodHttpStatus(t *testing.T) {
	server := statusServer()
	defer server.Close()

	bookmark := Bookmark{Href: server.URL + "/status/200"}
	checker := makeChecker()
	success, code, _ := checker.check(bookmark)
	if !success {
		t.Errorf("HTTP code %d should be treated as success", code)
	}
}

func TestCheckBadHttpStatus(t *testing.T) {
	server := statusServer()
	defer server.Close()

	bookmark := Bookmark{Href: server.URL + "/status/412"}
	checker := makeChecker()
	success, code, _ := checker.check(bookmark)
	if success {
		t.Errorf("HTTP code %d should be treated as failure", code)
	}
}

func TestSimpleReporterShowingAFailure(t *testing.T) {
	server := statusServer()
	defer server.Close()

	verbose := false
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		{Href: server.URL + "/status/404"},
		{Href: server.URL + "/status/200"},
	}
	checker := makeChecker()
	checker.Reporter = NewSimpleFailureReporter(verbose, true, &buffer)
	checker.Run(bookmarks)

	lineCount := strings.Count(buffer.String(), "\n")

	if lineCount != 1 {
		t.Errorf("One failure should have been reported, %d found", lineCount)
	}
}

func TestSimpleReporterShowingASucessInVerboseMode(t *testing.T) {
	server := statusServer()
	defer server.Close()

	verbose := true
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		{Href: server.URL + "/status/200"},
	}
	checker := makeChecker()
	checker.Reporter = NewSimpleFailureReporter(verbose, true, &buffer)
	checker.Run(bookmarks)

	lineCount := strings.Count(buffer.String(), "\n")

	if lineCount != 1 {
		t.Errorf("One success should have been reported, %d found", lineCount)
	}
}

func TestJSONReporterShowingAFailure(t *testing.T) {
	server := statusServer()
	defer server.Close()

	verbose := false
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		{Href: server.URL + "/status/404"},
		{Href: server.URL + "/status/200"},
	}

	reporter := NewJSONReporter(verbose, &buffer)

	checker := makeChecker()
	checker.Reporter = reporter

	checker.Run(bookmarks)

	failureCount := len(reporter.failures)
	if failureCount != 1 {
		t.Errorf("One failure should have been reported, %d found", failureCount)
	}

	failedBookmarks, err := ParseJSON(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("Unexpected error parsing JSON output: %s", err)
	}
	failedBookmarksCount := len(failedBookmarks)
	if failedBookmarksCount != 1 {
		t.Errorf("Expected one failed bookmark to be present in generated JSON, %d found", failedBookmarksCount)
	}

	failedBookmark := failedBookmarks[0]
	if failedBookmark.FailureInfo.HttpCode != 404 {
		t.Errorf("Wrong code set on failure info JSON, expected 404, got %d", failedBookmark.FailureInfo.HttpCode)
	}
}

func TestJSONReporterShowingSuccessInVerboseMode(t *testing.T) {
	server := statusServer()
	defer server.Close()

	verbose := true
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		{Href: server.URL + "/status/404"},
		{Href: server.URL + "/status/200"},
	}

	reporter := NewJSONReporter(verbose, &buffer)
	checker := makeChecker()
	checker.Reporter = reporter
	checker.Run(bookmarks)

	successCount := len(reporter.successes)
	if successCount != 1 {
		t.Errorf("One failure should have been reported, %d found", successCount)
	}

	failedBookmarks, err := ParseJSON(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("Unexpected error parsing JSON output: %s", err)
	}
	failedBookmarksCount := len(failedBookmarks)
	if failedBookmarksCount != 2 {
		t.Errorf("Expected two bookmarks to be present in generated JSON, %d found", failedBookmarksCount)
	}
}
