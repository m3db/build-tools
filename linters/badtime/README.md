# Badtime

Badtime is a Golang linter that detects inappropriate usage of the time.Time struct. Currently it detects the following:

1. Maps where the key of the map contains an instance of time.Time
2. Comparison of two time.Time structs using the == operator

While an instance of time.Time can be safely stored as part of the key of a map, it's very easy to introduce subtle bugs this way and it's often much safer to use something like an int64 to store a unix timestamp at nanosecond resolution instead.

Similarly, it CAN sometimes be correct to compare two time.Time structs using the `==` operator, however, its likely to lead to unintended behavior.

For detailed information on why storing a time.Time as a map key or comparing two time.Time structs using `==` can be dangerous, read [this section](https://golang.org/src/time/time.go?#L101) of the golang documentation for the time package.

Note that this issue can be particularly troublesome for projects upgrading to Golang 1.9 because in previous versions of Golang, two instances of time.Time for the same moment of time would only not be `==` to each other if they represented different timezones, whereas in Golang 1.9 and later they can also not be `==` to each other if one contains a monotonic bit and the other does not.

## Gometalinter integration

badtime is designed to integrate with [gometalinter](https://github.com/alecthomas/gometalinter). To add it to the list of active linters, make sure badtime is installed, and then modify the `.metalinter.json` file to add "badtime" to the "Enable" array and add the following entry in the "Linters" object: `"badtime": "badtime:PATH:LINE:COL:MESSAGE"`

Example:

```json
{
  "Linters": {
    "badtime": "badtime:PATH:LINE:COL:MESSAGE"
  },
  "Enable": ["badtime"],
}
```

## Installation

```bash
go get -u https://github.com/m3db/build-tools/linters/badtime
```

## Usage

```bash
badtime ./...
```

To view optional flags, run:

```bash
badtime -h
```

Note that the badtime package interprets path arguments and build tags the same way the standard Go toolchain does.

## Development

1. Clone this repo into your $GOPATH
2. [Make sure you have glide installed](https://github.com/Masterminds/glide)
3. Run glide install
4. Modify the code, add a new test file to the testdata directory, and update the testcases in main_test.go

### Running the tests

```bash
go test <PATH_TO_BADTIME_IN_YOUR_$GOPATH>
```