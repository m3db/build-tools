include .ci/common.mk

coverfile         := cover.out
test              := .ci/test-cover.sh

install: install-glide
	( cd linters/badtime ; glide install -v)
	( cd utilities/mockclean ; glide install -v)

test-internal:
	@which go-junit-report > /dev/null || go get -u github.com/sectioneight/go-junit-report
	$(test) $(coverfile) | tee $(test_log)

test-ci-unit: test-internal
	@which goveralls > /dev/null || go get -u -f github.com/mattn/goveralls
	goveralls -coverprofile=$(coverfile) -service=travis-ci || echo -e "\x1b[31mCoveralls failed\x1b[m"