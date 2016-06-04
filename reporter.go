package pinboardchecker

import (
	"fmt"
	"io"
	"os"
)

type SimpleFailureReporter struct {
	writers []io.Writer
}

func (r SimpleFailureReporter) onFailure(failure LookupFailure) {
	for _, writer := range r.writers {
		fmt.Fprintf(writer, "[ERR] %s\n", failure.Bookmark.Href)
	}
}

func (r SimpleFailureReporter) onSuccess(bookmark Bookmark) {
	for _, writer := range r.writers {
		fmt.Fprintf(writer, "[OK] %s\n", bookmark.Href)
	}
}

func newSimpleFailureReporter(writers ...io.Writer) SimpleFailureReporter {
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	return SimpleFailureReporter{writers: writers}
}
