package pinboard

import (
	"bytes"
	"strings"
	"testing"
)

func TestCheckGoodHttpStatus(t *testing.T) {
	bookmark := Bookmark{Href: "http://httpbin.org/html"}
	success, code, _ := check(bookmark)
	if !success {
		t.Errorf("HTTP code %d should be treated as success", code)
	}
}

func TestCheckBadHttpStatus(t *testing.T) {
	bookmark := Bookmark{Href: "http://httpbin.org/status/412"}
	success, code, _ := check(bookmark)
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
	checkAll(bookmarks, newSimpleFailureReporter(verbose, true, &buffer))

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
	checkAll(bookmarks, newSimpleFailureReporter(verbose, true, &buffer))

	lineCount := strings.Count(buffer.String(), "\n")

	if lineCount != 1 {
		t.Errorf("One success should have been reported, %d found", lineCount)
	}
}
