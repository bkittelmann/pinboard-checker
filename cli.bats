#!/usr/bin/env bats

# this endpoint will return the same content as in testdata/bookmarks.json
EXPORT_ENDPOINT="http://www.mocky.io/v2/5775832a0f0000e90997c48c/"

# simulates a successful delete via the pinboard API
DELETE_OK_ENDPOINT="http://www.mocky.io/v2/579755b6260000dd1217facc/"

# simulates a failed delete via the pinboard API
DELETE_FAIL_ENDPOINT="http://www.mocky.io/v2/579755d3260000dc1217facd/"

# this will prevent a token being accidentally read from your user's config
setup() {
	if [ -f ~/.pinboard-checker/pinboard-checker.yaml ]; then
		mv ~/.pinboard-checker/pinboard-checker.yaml ~/.pinboard-checker/bak.pinboard-checker.yaml
	fi 
}

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

@test "delete: Existing bookmark gets deleted" {
	run ./pinboard-checker delete -t "token" --endpoint $DELETE_OK_ENDPOINT http://example.com

	[ "$status" -eq 0 ]
}

@test "delete: Non-existing bookmark can not be deleted" {
	run ./pinboard-checker delete -t "token" --endpoint $DELETE_FAIL_ENDPOINT http://example.com

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

teardown() {
	if [ -f ~/.pinboard-checker/bak.pinboard-checker.yaml ]; then
		mv ~/.pinboard-checker/bak.pinboard-checker.yaml ~/.pinboard-checker/pinboard-checker.yaml
	fi 
}
