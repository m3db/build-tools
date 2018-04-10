DEVELOPER NOTES
===============

### Steps to add a new tool to this repository
(1) Create a new folder for the tool. E.g. if it's a new linter called `xyzcheck`, create `./linters/xyzcheck` and place all your code there.
(2) Add the target for the new tool in `Makefile` under the var decl `TARGETS := ...`
