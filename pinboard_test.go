package main

import (
	"os"
	"testing"
)

func TestParseJsonExport(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks := parseJson(file)

	if len(bookmarks) != 2 {
		t.Errorf("Expected 2 bookmark objects to be parsed from JSON, got %d", len(bookmarks))
	}

	nrOfTags := len(bookmarks[0].Tags)
	if nrOfTags != 4 {
		t.Errorf("Expected 4 tags, parse %d instead", nrOfTags)
	}
}
