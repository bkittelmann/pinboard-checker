package pinboardchecker

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
	var buffer bytes.Buffer

	bookmarks := []Bookmark{Bookmark{Href: "http://httpbin.org/status/404"}}
	checkAll(bookmarks, newSimpleFailureReporter(&buffer))

	lineCount := strings.Count(buffer.String(), "\n")

	if lineCount != 1 {
		t.Errorf("One failure should have been reported, %d found", lineCount)
	}
}

func TestSimpleReporterShowingASucessInVerboseMode(t *testing.T) {
	var buffer bytes.Buffer

	bookmarks := []Bookmark{Bookmark{Href: "http://httpbin.org/status/200"}}
	checkAll(bookmarks, newSimpleFailureReporter(&buffer))

	lineCount := strings.Count(buffer.String(), "\n")

	if lineCount != 1 {
		t.Errorf("One success should have been reported, %d found", lineCount)
	}
}
