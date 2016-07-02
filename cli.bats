#!/usr/bin/env bats

# this endpoint will return the same content as in testdata/bookmarks.json
EXPORT_ENDPOINT="http://www.mocky.io/v2/5775832a0f0000e90997c48c/"

@test "export: Get JSON output on stdout" {
	run ./pinboard-checker export -t 'token' --endpoint $EXPORT_ENDPOINT

	num_of_bookmarks=$(echo $output | jq length)

	[ "$status" -eq 0 ]
	[ $num_of_bookmarks = "2" ]
}

@test "export: Token argument is required" {
	run ./pinboard-checker export --endpoint $EXPORT_ENDPOINT

	[ "$status" -eq 1 ]
}