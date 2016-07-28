# pinboard-checker

[![Build Status](https://travis-ci.org/bkittelmann/pinboard-checker.svg?branch=master)](https://travis-ci.org/bkittelmann/pinboard-checker)
[![codecov](https://codecov.io/gh/bkittelmann/pinboard-checker/branch/master/graph/badge.svg)](https://codecov.io/gh/bkittelmann/pinboard-checker)

This command-line tool checks bookmarks stored on [pinboard.in](https://pinboard.in) if they are still working. For any errors encountered a report will be generated.

## Features

- Link lookup happens concurrently which makes generation of the final report fast
- Report of broken links can be shown on terminal or stored as JSON file
- Not tied to [pinboard.in](https://pinboard.in), can be used to check any list of URLs given as input
- Various configuration options to fine-tune performance of link lookups
- Separate command to export all of your bookmarks
- Allows deletion of specific bookmarks
- Can be used as library to interact with [pinboard.in](https://pinboard.in/api) API

## Installation

```
go get github.com/bkittelmann/pinboard-checker
```

## Usage

The core functionality of this tool is to check a list of bookmarks, whether each bookmark's URL can be resolved via HTTP. If an error is encountered, it will be collected along with auxiliary information that might help to point to the reason why a URL failed (e.g. a HTTP error code of 404, or a DNS lookup error if an address does not exist anymore). Failed URLs will be reported.

`pinboard-checker` will not change nor delete the bookmarks while the URL lookup ("check") operation is still running. This is because it is left up to the user to decide what to do with a stale URL. Some errors may be temporary, other URLs might be invalid since a long time.

The `pinboard-checker` application can show in a help dialog which commands are supported:

```
$ ./pinboard-checker help
Tool for checking the state of your links on pinboard.in

Usage:
  pinboard-checker [command]

Available Commands:
  check       Check for stale links
  delete      Bulk-delete links stored in your pinboard
  export      Download your bookmarks

Flags:
      --endpoint string   URL of pinboard API endpoint (default "https://api.pinboard.in")
  -t, --token string      The pinboard API token

Use "pinboard-checker [command] --help" for more information about a command.
```

Common to all command is that you have to supply the `--token` flag. This contains your pinboard API token which you can find on the [settings/password](https://pinboard.in/settings/password) page in your pinboard account. This token is used to read and edit your bookmarks.

### `check` command

Use `check` to iterate through the list of your bookmarks and report any errors:

```
$ ./pinboard-checker check -t APITOKEN
[ERR] http://httpbin.org/status/404 HTTP status: 404
```

By default this will connect to your pinboard account, read all your bookmarks, and check them all.

### `delete` command

Easily delete URLs that you have bookmarked.

```
$ ./pinboard-checker delete -t APITOKEN http://example.com
```

You can either supply the URLs to be deleted as arguments to the `delete` command, or read the content of a file (see the `--inputFile` parameter documentation).

### `export` command

Exports all your bookmarks and writes them directly on `stdout`. Use standard redirection to save output in a file.

```
/pinboard-checker export -t APITOKEN > backup_bookmarks.json
```

## Development notes

### Running unit tests

```
go test -v './...'
```

###  Testing the program via bash

As a form of integration test, we use the [`Bats`](https://github.com/sstephenson/bats#readme)
framework to test the commandline interface of `pinboard-checker`. Run the test like this:

```
$ bats cli.bats
 ✓ check: Token argument is required
 ✓ check: Timeout flag restricts runtime of link lookup
 ✓ delete: Token argument is required
 ✓ export: Get JSON output on stdout
 ✓ export: Token argument is required

5 tests, 0 failures
```
