# pinboard-checker

[![Build Status](https://travis-ci.org/bkittelmann/pinboard-checker.svg?branch=master)](https://travis-ci.org/bkittelmann/pinboard-checker)

Checks the bookmarks stored on pinboard.in if they are still working. 

## Features

- Link lookup happens concurrently which makes generation of the final report fast
- Report of broken links can be shown on terminal or stored as JSON file
- Not tied to pinboard.in, can be used to check any list of URLs given as input
- Various configuration options to fine-tune performance of link lookups
- Separate command to export all of your bookmarks
- Allows deletion of specific bookmarks

## Installation

```
go get github.com/bkittelmann/pinboard-checker
```

##  Testing the program via bash

As a kind of integration test, we use the [`Bats`](https://github.com/sstephenson/bats#readme)
framework to test the commandline interface of `pinboard-checker`. Run the test like this:

```
$ bats cli.bats
 ✓ export: Get JSON output on stdout
 ✓ export: Token argument is required

2 tests, 0 failures
```