package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type SimpleFailureReporter struct {
	writers []io.Writer
	verbose bool
}

func (r SimpleFailureReporter) constructErrorMessage(failure LookupFailure) string {
	if failure.Code > 0 {
		return fmt.Sprintf("HTTP status: %d", failure.Code)
	}
	errorParts := strings.Split(failure.Error.Error(), ": ")
	return fmt.Sprintf("Other: %s", errorParts[len(errorParts)-1])
}

func (r SimpleFailureReporter) onFailure(failure LookupFailure) {
	for _, writer := range r.writers {
		fmt.Fprintf(writer, "[ERR] %s %s\n", failure.Bookmark.Href, r.constructErrorMessage(failure))
	}
}

func (r SimpleFailureReporter) onSuccess(bookmark Bookmark) {
	if r.verbose {
		for _, writer := range r.writers {
			fmt.Fprintf(writer, "[OK] %s\n", bookmark.Href)
		}
	}
}

func newSimpleFailureReporter(verbose bool, writers ...io.Writer) SimpleFailureReporter {
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	return SimpleFailureReporter{verbose: verbose, writers: writers}
}
