package pinboard

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks := parseJSON(file)

	if len(bookmarks) != 2 {
		t.Errorf("Expected 2 bookmark objects to be parsed from JSON, got %d", len(bookmarks))
	}

	nrOfTags := len(bookmarks[0].Tags)
	if nrOfTags != 4 {
		t.Errorf("Expected 4 tags, parse %d instead", nrOfTags)
	}
}

func TestWriteJSON(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks := parseJSON(file)

	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	writeJSON(bookmarks, buf)
	buf.Flush()

	deserialized := parseJSON(bufio.NewReader(&b))

	if !reflect.DeepEqual(bookmarks, deserialized) {
		t.Errorf("Deserialization did not work")
	}
}