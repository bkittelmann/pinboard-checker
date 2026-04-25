#!/usr/bin/env bats

# Build the local mock server once per file and start it in the background.
# Tests reach it via $MOCK_URL.
setup_file() {
	export MOCK_BIN="$BATS_FILE_TMPDIR/mockserver"
	go build -o "$MOCK_BIN" ./testdata/mockserver

	"$MOCK_BIN" > "$BATS_FILE_TMPDIR/mock_url" &
	echo $! > "$BATS_FILE_TMPDIR/mock_pid"

	# Wait for the server to print its URL.
	for _ in $(seq 1 50); do
		[ -s "$BATS_FILE_TMPDIR/mock_url" ] && break
		sleep 0.1
	done

	export MOCK_URL=$(cat "$BATS_FILE_TMPDIR/mock_url")
	export EXPORT_ENDPOINT="$MOCK_URL/export/"
	export DELETE_OK_ENDPOINT="$MOCK_URL/delete-ok/"
	export DELETE_FAIL_ENDPOINT="$MOCK_URL/delete-fail/"
	export DELAY_URL="$MOCK_URL/delay/3"
}

teardown_file() {
	if [ -f "$BATS_FILE_TMPDIR/mock_pid" ]; then
		kill "$(cat "$BATS_FILE_TMPDIR/mock_pid")" 2>/dev/null || true
	fi
}

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

	output=$(time (echo "$DELAY_URL" | ./pinboard-checker check -i - --inputFormat=txt --timeout=1s 2>/dev/null 1>&2) 2>&1)

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
