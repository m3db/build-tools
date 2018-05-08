# ImportOrder

ImportOrder is a Golang linter that detects misordered imports given a certain pattern set.

For example, let's say you run the following `importorder` command:

./importorder -patterns="STDLIB github.com/m3db/m3coordinator github.com/m3db EXTERNAL" path/to/directory

This will make sure that your import groups (in `path/to/directory`) start with all Go standard library imports, followed by imports in the package github.com/m3db/m3coordinator, followed by all imports in the github.com/m3db package, and finally all external/third party packages.

In other words, this block will succeed:
```
import (
	"context"
	"time"

	"github.com/m3db/m3coordinator/errors"
	"github.com/m3db/m3coordinator/policy/resolver"

	"github.com/m3db/m3db/client"
	"github.com/m3db/m3x/ident"
)
```

And this will fail because `context` should belong in the first group with the other standard library import:
```
import (
	"time"

	"github.com/m3db/m3coordinator/errors"
	"context"
	"github.com/m3db/m3coordinator/policy/resolver"

	"github.com/m3db/m3db/client"
	"github.com/m3db/m3x/ident"
)
```

There are a few notes to point out:

1. If you are going to have two patterns where one is a subset of the other (e.g. `github.com/m3db/m3coordinator` and `github.com/m3db`), make sure that you provide the more specific one first. Otherwise, the linter will provide inaccurate results.
2. If you want to see exactly how the imports should look like as opposed to just getting an error, set the `verbose` flag to `true` (e.g. `./importorder -patterns="STDLIB github.com/m3db/m3coordinator github.com/m3db EXTERNAL" -verbose=true path/to/directory`)
3. If you want to specify Go's standard library imports, use "STDLIB", and if you want to have a catch-all, use "EXTERNAL" (for all other third party/external packages)

## Gometalinter integration

`importorder` is designed to integrate with [gometalinter](https://github.com/alecthomas/gometalinter). To add it to the list of active linters, make sure `importorder` is installed, and then modify the `.metalinter.json` file to add "importorder" to the "Enable" array and also add it to the "Linters" object.

Example:

```json
{
  "Linters": {
    "importorder": {
      "Command": "importorder -patterns=\"STDLIB github.com/m3db/m3coordinator github.com/m3db EXTERNAL\"",
      "Pattern": "PATH:LINE:MESSAGE"
    },
  },
  "Enable":
    [ "importorder" ],
}
```

## Installation

```bash
go get -u github.com/m3db/build-tools/linters/importorder
cd $GOPATH/src/github.com/m3db/build-tools/linters/importorder
glide install -v
go install .
```

## Usage

```bash
importorder -patterns="pattern_1 pattern_2 pattern_3 pattern_4" ./...
```

To view optional flags, run:

```bash
importorder -h
```

Note that the importorder package interprets path arguments and build tags the same way the standard Go toolchain does.

## Development

1. Clone this repo into your $GOPATH
2. [Make sure you have glide installed](https://github.com/Masterminds/glide)
3. Run glide install
4. Modify the code, add a new test file to the testdata directory, and update the testcases in main_test.go

### Running the tests

```bash
go test <PATH_TO_IMPORTORDER_IN_YOUR_$GOPATH>
```