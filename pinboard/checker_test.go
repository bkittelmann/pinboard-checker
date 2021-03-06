package pinboard

import (
	"bytes"
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

func TestCheckGoodHttpStatus(t *testing.T) {
	bookmark := Bookmark{Href: "http://httpbin.org/html"}
	checker := makeChecker()
	success, code, _ := checker.check(bookmark)
	if !success {
		t.Errorf("HTTP code %d should be treated as success", code)
	}
}

func TestCheckBadHttpStatus(t *testing.T) {
	bookmark := Bookmark{Href: "http://httpbin.org/status/412"}
	checker := makeChecker()
	success, code, _ := checker.check(bookmark)
	if success {
		t.Errorf("HTTP code %d should be treated as failure", code)
	}
}

func TestSimpleReporterShowingAFailure(t *testing.T) {
	verbose := false
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		Bookmark{Href: "http://httpbin.org/status/404"},
		Bookmark{Href: "http://httpbin.org/status/200"},
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
	verbose := true
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		Bookmark{Href: "http://httpbin.org/status/200"},
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
	verbose := false
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		Bookmark{Href: "http://httpbin.org/status/404"},
		Bookmark{Href: "http://httpbin.org/status/200"},
	}

	reporter := NewJSONReporter(verbose, &buffer)

	checker := makeChecker()
	checker.Reporter = reporter

	checker.Run(bookmarks)

	failureCount := len(reporter.failures)
	if failureCount != 1 {
		t.Errorf("One failure should have been reported, %d found", failureCount)
	}

	failedBookmarks := ParseJSON(bytes.NewReader(buffer.Bytes()))
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
	verbose := true
	var buffer bytes.Buffer

	bookmarks := []Bookmark{
		Bookmark{Href: "http://httpbin.org/status/404"},
		Bookmark{Href: "http://httpbin.org/status/200"},
	}

	reporter := NewJSONReporter(verbose, &buffer)
	checker := makeChecker()
	checker.Reporter = reporter
	checker.Run(bookmarks)

	successCount := len(reporter.successes)
	if successCount != 1 {
		t.Errorf("One failure should have been reported, %d found", successCount)
	}

	failedBookmarks := ParseJSON(bytes.NewReader(buffer.Bytes()))
	failedBookmarksCount := len(failedBookmarks)
	if failedBookmarksCount != 2 {
		t.Errorf("Expected two bookmarks to be present in generated JSON, %d found", failedBookmarksCount)
	}
}
