#!/usr/bin/env bats

# this endpoint will return the same content as in testdata/bookmarks.json
EXPORT_ENDPOINT="http://www.mocky.io/v2/5775832a0f0000e90997c48c/"

@test "check: Token argument is required" {
	run ./pinboard-checker check

	[ "$status" -eq 1 ]
}

@test "check: Timeout flag restricts runtime of link lookup" {
	# Note: The `time` command outputs on stderr, won't be captured by bats generally.
	# That's why we have to redirect the output into a variable to match its content.

	output=$(time (echo 'http://httpbin.org/delay/3' | ./pinboard-checker check -i - --inputFormat=txt --timeout=1s 2>/dev/null 1>&2) 2>&1)
	
	# this checks that the time taken is 1 second and a few miliseconds
	[[ $output =~ real[[:space:]]0m1.[0-9]+s ]]	
}

@test "delete: Token argument is required" {
	run ./pinboard-checker delete

	[ "$status" -eq 1 ]
}

@test "export: Get JSON output on stdout" {
	run ./pinboard-checker export -t 'token' --endpoint $EXPORT_ENDPOINT

	num_of_bookmarks=$(echo $output | jq length)

	[ "$status" -eq 0 ]
	[ $num_of_bookmarks = "2" ]
}

@test "export: Token argument is required" {
	run ./pinboard-checker export

	[ "$status" -eq 1 ]
}

