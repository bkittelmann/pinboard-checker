// Mock server used by cli.bats integration tests.
//
// It serves three pinboard-API-shaped endpoints (export, delete-ok,
// delete-fail) and a /delay/<seconds> handler used to exercise the
// check command's --timeout flag. The server prints its base URL on
// stdout and runs until it receives SIGINT/SIGTERM.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const bookmarksJSON = `[
  {
    "href": "https://blog.bracelab.com/achieving-perfect-ssl-labs-score-with-go",
    "description": "Achieving a Perfect SSL Labs Score with Go",
    "extended": "",
    "meta": "a9138eea3d5ea2cf7dedf29c70a9b786",
    "hash": "811a463298eeb77d7e3ddb2119d4b159",
    "time": "2016-05-29T10:16:11Z",
    "shared": "yes",
    "toread": "no",
    "tags": "golang http ssl configuration"
  },
  {
    "href": "https://github.com/xenolf/lego",
    "description": "xenolf/lego: Let's Encrypt client and ACME library written in Go",
    "extended": "",
    "meta": "4ad7b349071a3df9b96b5dabe4dddf04",
    "hash": "6f02cfface77e59cbce7788d44095fcc",
    "time": "2016-05-29T10:14:53Z",
    "shared": "no",
    "toread": "yes",
    "tags": "golang encryption library tool ssl certificates"
  }
]`

func canned(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}
}

func main() {
	mux := http.NewServeMux()

	// Endpoints used by the bats tests. Each one is a base URL the binary
	// will append paths like /v1/posts/all or /v1/posts/delete to; the
	// canned response is returned regardless of the suffix.
	mux.Handle("/export/", canned(bookmarksJSON))
	mux.Handle("/delete-ok/", canned(`{"result_code":"done"}`))
	mux.Handle("/delete-fail/", canned(`{"result_code":"item not found"}`))

	// /delay/<seconds> sleeps and returns 200, used by the timeout test.
	mux.HandleFunc("/delay/", func(w http.ResponseWriter, r *http.Request) {
		secs := strings.TrimPrefix(r.URL.Path, "/delay/")
		d, err := time.ParseDuration(secs + "s")
		if err != nil {
			http.Error(w, "bad delay", http.StatusBadRequest)
			return
		}
		time.Sleep(d)
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	fmt.Println(server.URL)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
