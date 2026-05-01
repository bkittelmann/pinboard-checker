package pinboard

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDeleteNonExistingBookmarkReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"result_code":"item not found"}`)
	}))
	defer server.Close()

	endpointUrl, _ := url.Parse(server.URL)

	client := NewClient("token", endpointUrl)

	err := client.DeleteBookmark(Bookmark{Href: "example.com", Description: ""})
	if err == nil {
		t.Error("Expected an error to be returned if non-existing URL is being deleted")
	}
}

func TestDeleteExistingBookmark(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"result_code":"done"}`)
	}))
	defer server.Close()

	endpointUrl, _ := url.Parse(server.URL)

	client := NewClient("token", endpointUrl)

	err := client.DeleteBookmark(Bookmark{Href: "example.com", Description: ""})
	if err != nil {
		t.Errorf("No error expected, got %s", err)
	}
}

func TestParseJSON(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks, err := ParseJSON(file)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if len(bookmarks) != 2 {
		t.Errorf("Expected 2 bookmark objects to be parsed from JSON, got %d", len(bookmarks))
	}

	nrOfTags := len(bookmarks[0].Tags)
	if nrOfTags != 4 {
		t.Errorf("Expected 4 tags, parse %d instead", nrOfTags)
	}
}

func TestParseJSONReturnsErrorOnBadInput(t *testing.T) {
	_, err := ParseJSON(strings.NewReader("not json"))
	if err == nil {
		t.Error("Expected an error for malformed JSON input")
	}
}

func TestWriteJSON(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks, err := ParseJSON(file)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	writeJSON(bookmarks, buf)
	buf.Flush()

	deserialized, err := ParseJSON(bufio.NewReader(&b))
	if err != nil {
		t.Fatalf("Unexpected error deserializing: %s", err)
	}

	if !reflect.DeepEqual(bookmarks, deserialized) {
		t.Errorf("Deserialization did not work")
	}
}

func TestTxtInputFormatForReadingFromFile(t *testing.T) {
	inputFile := strings.NewReader(`
		http://example.com/a
		http://example.com/b
	`)
	bookmarks, err := GetBookmarksFromFile(inputFile, TXT)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(bookmarks) != 2 {
		t.Errorf("Text links were not parsed as bookmarks for input")
	}
}

func TestJSONInputFormatForReadingFromFile(t *testing.T) {
	file, _ := os.Open("testdata/bookmarks.json")
	defer file.Close()

	bookmarks, err := GetBookmarksFromFile(file, JSON)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(bookmarks) != 2 {
		t.Errorf("JSON links were not parsed as bookmarks for input")
	}
}
