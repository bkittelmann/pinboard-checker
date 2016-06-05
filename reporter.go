package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type SimpleFailureReporter struct {
	writers        []io.Writer
	verbose        bool
	colorizePrefix bool
}

func (r SimpleFailureReporter) makeSuccessPrefix() string {
	prefix := "[OK] "
	if r.colorizePrefix {
		return color.New(color.FgGreen).SprintFunc()(prefix)
	}
	return prefix
}

func (r SimpleFailureReporter) makeFailurePrefix() string {
	prefix := "[ERR] "
	if r.colorizePrefix {
		return color.New(color.FgRed).SprintFunc()(prefix)
	}
	return prefix
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
		fmt.Fprintf(writer, "%s%s %s\n", r.makeFailurePrefix(), failure.Bookmark.Href, r.constructErrorMessage(failure))
	}
}

func (r SimpleFailureReporter) onSuccess(bookmark Bookmark) {
	if r.verbose {
		for _, writer := range r.writers {
			fmt.Fprintf(writer, "%s%s\n", r.makeSuccessPrefix(), bookmark.Href)
		}
	}
}

func newSimpleFailureReporter(verbose bool, colorize bool, writers ...io.Writer) SimpleFailureReporter {
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	return SimpleFailureReporter{
		writers:        writers,
		verbose:        verbose,
		colorizePrefix: colorize,
	}
}
