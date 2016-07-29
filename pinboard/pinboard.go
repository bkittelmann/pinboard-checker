package pinboard

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

var logger = logrus.New()

type Format int

const (
	JSON Format = iota + 1
	TXT
)

func (f Format) String() string {
	if f == JSON {
		return "json"
	}
	if f == TXT {
		return "txt"
	}
	return ""
}

func FormatFromString(value string) (Format, error) {
	switch value {
	case "json":
		return JSON, nil
	case "txt":
		return TXT, nil
	}
	return 0, errors.New(fmt.Sprintf("%s is not a valid format value", value))
}

type PinboardBoolean bool

func (p *PinboardBoolean) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*p = (value == "yes")
	return nil
}

func (p *PinboardBoolean) MarshalJSON() ([]byte, error) {
	if *p {
		return json.Marshal("yes")
	} else {
		return json.Marshal("no")
	}
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

func (p *PinboardTags) MarshalJSON() ([]byte, error) {
	out := strings.Join(*p, " ")
	result, err := json.Marshal(out)
	return result, err
}

type FailureInfo struct {
	HttpCode     int    `json:"httpCode,omitempty"`
	ErrorMessage string `json:"message,omitempty"`
	// note: needs to be a pointer type so that 'omitempty' does work
	CheckedAt *time.Time `json:"checkedAt,omitempty"`
}

type Bookmark struct {
	Href        string          `json:"href"`
	Description string          `json:"description,omitempty"`
	Extended    string          `json:"extended,omitempty"`
	Meta        string          `json:"meta,omitempty"`
	Hash        string          `json:"hash,omitempty"`
	Time        time.Time       `json:"time,omitempty"`
	Shared      PinboardBoolean `json:"shared"`
	ToRead      PinboardBoolean `json:"toread"`
	Tags        PinboardTags    `json:"tags"`
	FailureInfo FailureInfo     `json:"failure,omitempty"`
}

func ParseJSON(input io.Reader) []Bookmark {
	var bookmarks []Bookmark
	json.NewDecoder(input).Decode(&bookmarks)
	return bookmarks
}

func ParseText(input io.Reader) []Bookmark {
	var bookmarks []Bookmark
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if len(url) > 0 {
			bookmarks = append(bookmarks, Bookmark{Href: url})
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}
	return bookmarks
}

func writeJSON(bookmarks []Bookmark, output io.Writer) {
	json.NewEncoder(output).Encode(bookmarks)
}

var DefaultEndpoint *url.URL

func init() {
	url, err := url.Parse("https://api.pinboard.in")
	if err == nil {
		DefaultEndpoint = url
	}
}

type Client struct {
	Token    string
	Endpoint *url.URL
}

func (client *Client) buildDownloadEndpoint() string {
	downloadPath, _ := url.Parse("v1/posts/all?format=json&auth_token=" + client.Token)
	endpoint := client.Endpoint.ResolveReference(downloadPath)
	return endpoint.String()
}

func (client *Client) buildDeleteEndpoint(rawUrl string) string {
	downloadPath, _ := url.Parse("v1/posts/delete?format=json&auth_token=" + client.Token)
	endpoint := client.Endpoint.ResolveReference(downloadPath)
	query := endpoint.Query()
	query.Add("url", rawUrl)
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}

func (client *Client) DownloadBookmarks() (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", client.buildDownloadEndpoint(), nil)

	httpClient := &http.Client{}
	response, err := httpClient.Do(req)

	if err != nil {
		logger.Debugf("Error %s", err)
		return nil, err
	}

	return response.Body, err
}

func (client *Client) GetAllBookmarks() ([]Bookmark, error) {
	readCloser, err := client.DownloadBookmarks()
	defer readCloser.Close()
	return ParseJSON(readCloser), err
}

func (client *Client) DeleteBookmark(bookmark Bookmark) (err error) {
	endpoint := client.buildDeleteEndpoint(bookmark.Href)

	logger.Debugf("Deleting %s\n", bookmark.Href)

	response, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// anonymous struct for response, TODO: Make it a type Result
	result := struct {
		Code string `json:"result_code"`
	}{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if result.Code == "item not found" {
		return errors.New(fmt.Sprintf("%s was not found in pinboard", bookmark.Href))
	}

	if result.Code != "done" {
		return errors.New(fmt.Sprintf("unexpected result code '%s'", result.Code))
	}

	return err
}

func NewClient(token string, endpoint *url.URL) *Client {
	return &Client{Token: token, Endpoint: endpoint}
}

func GetBookmarksFromFile(reader io.Reader, format Format) []Bookmark {
	var bookmarks []Bookmark
	switch format {
	case TXT:
		bookmarks = ParseText(reader)
	case JSON:
		bookmarks = ParseJSON(reader)
	}
	return bookmarks
}
