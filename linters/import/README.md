# Import

Import is a Golang linter that detects misordered imports given a certain pattern set.

For example, let's say you have the following pattern set for your imports:

```
standard
github.com/m3db/m3coordinator
github.com/m3db
ext
```

This will make sure that your import groups start with all Go standard library imports, followed by imports in the package github.com/m3db/m3coordinator, followed by all imports in the github.com/m3db package, and finally all external/third party packages.

There are a few notes to point out:

1. If you are going to have two patterns where one is a subset of the other (e.g. `github.com/m3db/m3coordinator` and `github.com/m3db`), make sure that you provide the more specific one first. Otherwise, the linter will provide inaccurate results.
2. Occasionally, you will get the error `number of import groups exceeds number of patterns. this usually means a previous or future import is out of order or you didn't specify the correct patterns`. Once we run out of patterns to check (i.e there are more import groups than patterns remaining), we stop checking and return. In this case, you should fix the imports for that file and rerun the linter to make sure no errors remain.
3. If you want to specify Go's standard library imports, use "standard", and if you want to have a catch-all, use "ext" (for all other third party/external packages)

## Installation

```bash
go get -u github.com/m3db/build-tools/linters/import
```

## Usage

```bash
import -patterns="pattern_1 pattern_2 pattern_3 pattern_4" ./...
```

To view optional flags, run:

```bash
import -h
```

Note that the import package interprets path arguments and build tags the same way the standard Go toolchain does.

## Development

1. Clone this repo into your $GOPATH
2. [Make sure you have glide installed](https://github.com/Masterminds/glide)
3. Run glide install
4. Modify the code, add a new test file to the testdata directory, and update the testcases in main_test.go

### Running the tests

```bash
go test <PATH_TO_IMPORT_IN_YOUR_$GOPATH>
```