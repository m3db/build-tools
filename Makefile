SELF_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include $(SELF_DIR)/.ci/common.mk

coverfile := cover.out
test      := .ci/test-cover.sh

TARGETS :=             \
	linters/badtime      \
	utilities/genclean  \
	linters/importorder  \
	utilities/ggd        \

define TARGET_RULES

.PHONY: setup_$(TARGET)
setup_$(TARGET):
	@echo Installing deps for $(TARGET)
	cd $(TARGET) && glide install -v

.PHONY: test_$(TARGET)
test_$(TARGET): setup_$(TARGET)
	@echo Building $(TARGET)
	cd $(TARGET) && (                                                                \
		go test -v -race -timeout 5m -covermode atomic -coverprofile $(coverfile) . && \
			$(SELF_DIR)/$(codecov_push) -f $(coverfile) &&                               \
		[ ! -f test.sh ] || ./test.sh                                                  \
	)

.PHONY: build_$(TARGET)
build_$(TARGET): setup_$(TARGET)
	cd $(TARGET) && go build .

.PHONY: $(TARGET)
$(TARGET): test_$(TARGET) build_$(TARGET)

endef

$(foreach TARGET,$(TARGETS),$(eval $(TARGET_RULES)))

.PHONY: targets
targets: $(TARGETS)

.PHONY: all
all: targets

.DEFAULT_GOAL := all
