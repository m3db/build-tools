GenClean
========

`genclean` is a utility to address a few shortcomings in `mockgen` (and other code-gen tools). Specifically:

- Remove self referential imports;
- Offers determinism when import aliasing;
- Grouping imports into chunks based on prefixes;

## Installation

```sh
go get -u github.com/m3db/build-tools/utilities/genclean
```

## Usage

The following command orders the imports into three chunks: stdlib, those starting with "github.com/some", and third-party; and aliases them deterministically; and removes any imports from the same package.

```sh
mockgen -package=abc github.com/some/path/abc IFace0 \
 | genclean -cleanup-selfref -cleanup-import -prefixes "github.com/some" -pkg github.com/some/path/abc -out $GOPATH/src/github.com/some/path/abc/abc_mock.go
```

You can embed this inside a `go:generate` command, as follows:

```go
//go:generate sh -c "mockgen -package=abc github.com/some/path/abc IFace0 | genclean ..."
```

## Known Limitations
- `genclean` assumes all go files in a folder `$GOPATH/src/github.com/xyz/abc` have package name set to the name of the folder ("abc" in this case).
