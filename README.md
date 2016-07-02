# Testing the program via bash

As a kind of integration test, we use the [`Bats`](https://github.com/sstephenson/bats#readme)
framework to test the commandline interface of `pinboard-checker`. Run the test like this:

```
$ bats cli.bats
 ✓ export: Get JSON output on stdout
 ✓ export: Token argument is required

2 tests, 0 failures
```