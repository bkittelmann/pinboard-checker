package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Bookmark struct {
	Href        string
	Description string
}

func parseJson(bookmarkJson []byte) []Bookmark {
	var bookmarks []Bookmark
	json.Unmarshal(bookmarkJson, &bookmarks)
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

func downloadBookmarks(token string) ([]byte, error) {
	response, err := http.Get(buildDownloadEndpoint(token))
	defer response.Body.Close()

	if err != nil {
		debug("Error %s", err)
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

func getAllBookmarks(token string) ([]Bookmark, error) {
	bookmarkJson, err := downloadBookmarks(token)
	return parseJson(bookmarkJson), err
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
