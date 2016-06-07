package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PinboardBoolean bool

func (p *PinboardBoolean) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*p = (value == "yes")
	return nil
}

type PinboardTags []string

func (p *PinboardTags) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = strings.Split(value, " ")
	return nil
}

type Bookmark struct {
	Href        string
	Description string
	Extended    string
	Meta        string
	Hash        string
	Time        time.Time
	Shared      PinboardBoolean
	ToRead      PinboardBoolean
	Tags        PinboardTags
}

func parseJson(input io.Reader) []Bookmark {
	var bookmarks []Bookmark
	json.NewDecoder(input).Decode(&bookmarks)
	return bookmarks
}

func buildDownloadEndpoint(token string) string {
	endpoint, _ := url.Parse("https://api.pinboard.in/v1/posts/all")
	query := endpoint.Query()
	query.Add("auth_token", token)
	query.Add("format", "json")
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func buildDeleteEndpoint(token string, rawUrl string) string {
	endpoint, _ := url.Parse("https://api.pinboard.in/v1/posts/delete")
	query := endpoint.Query()
	query.Add("auth_token", token)
	query.Add("format", "json")
	query.Add("url", rawUrl)
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func downloadBookmarks(token string) (io.ReadCloser, error) {
	response, err := http.Get(buildDownloadEndpoint(token))

	if err != nil {
		debug("Error %s", err)
		return nil, err
	}

	return response.Body, err
}

func getAllBookmarks(token string) ([]Bookmark, error) {
	readCloser, err := downloadBookmarks(token)
	defer readCloser.Close()
	return parseJson(readCloser), err
}

func deleteBookmark(token string, bookmark Bookmark) {
	endpoint := buildDeleteEndpoint(token, bookmark.Href)

	debug("Deleting %s\n", bookmark.Href)

	response, err := http.Get(endpoint)
	defer response.Body.Close()

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	debug("%s", body)
}
