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
		$(SELF_DIR)/$(codecov_push) -f $(coverfile) -F $(TARGET)

endef

$(foreach TARGET,$(TARGETS),$(eval $(TARGET_RULES)))

.PHONY: targets
targets: $(TARGETS)

.PHONY: all
all: targets

.DEFAULT_GOAL := all
