ggd - GoGitDiff
===============

This is a tool to compute the packages affected by the difference between two Git
revisions. It's primary goal is to help speed up CI jobs by figuring out the packages
affected by git changes, and transitively looking through the Go Imports. It takes
inspiration from `gta`, and the work done at [DigitalOcean].

NB: debug mode (`-d`) allows users to inspect how the tool arrives at the decisions it does.

  [DigitalOcean]: https://blog.digitalocean.com/cthulhu-organizing-go-code-in-a-scalable-repo/

### Installation
```sh
$ git checkout github.com/m3db/build-tools
$ cd build-tools/utilities/ggd
$ glide install -v
$ go install .
```

### Sample Usage
```sh
ggd: command line tool to find packages affected by git changes. Examples:
# Assuming CWD is in a git repository directory present in the GOPATH.

# (1) List all the golang packages affected by changes between master and HEAD
ggd

# (2) List all the golang packages affected by changes between branchA and HEAD
ggd branchA

# (3) List all the golang packages affected by between changes between branchA and branchB
ggd branchA..branchB

# (4) Run tests for all the golang packages affected by between master, and head
go test $(ggd)

# (5) List all the golang packages affected by between master and head, and build tags 'integration'
ggd -t integration

# (6) The same as (5), but include debug output (sent to STDERR)
ggd -t integration -d

# (7) The same as (6), but include debug output (sent to STDERR), and
# save the generated DAG in changes.png for visualization
ggd -t integration -d -o change.dot
dot -Tpng change.png change.dot
```
