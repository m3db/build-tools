SELF_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include $(SELF_DIR)/.ci/common.mk

coverfile         := cover.out
test              := .ci/test-cover.sh

TARGETS :=             \
	linters/badtime      \
	utilities/mockclean  \

define TARGET_RULES

.PHONY: setup_$(TARGET)
setup_$(TARGET):
	@echo Installing deps for $(TARGET)
	cd $(TARGET) && glide install -v

.PHONY: $(TARGET)
$(TARGET): setup_$(TARGET)
	@which go-junit-report > /dev/null || go get -u github.com/sectioneight/go-junit-report
	@echo Building $(TARGET)
	cd $(TARGET) && go test -v -race -timeout 5m -covermode atomic -coverprofile $(coverfile) ./... && \
		goveralls -coverprofile=$(coverfile) -service=travis-ci || (echo -e "\x1b[31mCoveralls failed\x1b[m" && exit 1)

endef

$(foreach TARGET,$(TARGETS),$(eval $(TARGET_RULES)))

.PHONY: targets
targets: $(TARGETS)

.PHONY: all
all: targets

.DEFAULT_GOAL := all
