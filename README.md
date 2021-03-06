# build-tools [![Build Status][ci-img]][ci] [![Coverage Status](https://codecov.io/gh/m3db/build-tools/branch/master/graph/badge.svg)](https://codecov.io/gh/m3db/build-tools)

Various build tools used as part of the M3DB project

[ci-img]: https://semaphoreci.com/api/v1/m3db/build-tools/branches/master/shields_badge.svg
[ci]: https://semaphoreci.com/m3db/build-tools

## Contents

1. [badtime](https://github.com/m3db/build-tools/blob/master/linters/badtime/README.md) - gometalinter plugin that detects inappropriate usage of the time.Time struct.
2. [genclean](https://github.com/m3db/build-tools/blob/master/utilities/genclean/README.md) - CLI tool to make `mockgen` runs idempotent.
3. [importorder](https://github.com/m3db/build-tools/blob/master/linters/importorder/README.md) - gometalinter plugin that detects accuracy of import ordering based on user specified patterns.
4. [ggd](https://github.com/m3db/build-tools/blob/master/utilities/ggd/README.md) - CLI tool to find which go packages are affected by git changes. Useful to speedup CI builds.

<hr>

This project is released under the [Apache License, Version 2.0](LICENSE).
