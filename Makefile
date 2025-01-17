# tool marcros
SHELL := /bin/bash
CC := go
CCFLAG :=
# path marcros
BIN_PATH := dist
OBJ_PATH := obj
SRC_PATH := cmd

# compile marcros
TARGET_NAME := google-cloudflare-sync
ifeq ($(OS),Windows_NT)
	TARGET_NAME := $(addsuffix .exe,$(TARGET_NAME))
endif
TARGET := $(BIN_PATH)/$(TARGET_NAME)
MAIN_SRC := cmd/$(TARGET_NAME)/main.go

# src files & obj files
SRC := $(foreach x, $(SRC_PATH), $(wildcard $(addprefix $(x)/*,.c*)))
OBJ := $(addprefix $(OBJ_PATH)/, $(addsuffix .o, $(notdir $(basename $(SRC)))))

# clean files list
DISTCLEAN_LIST := $(OBJ)
CLEAN_LIST := $(TARGET) \
			  $(DISTCLEAN_LIST)

# default rule
default: all

# non-phony targets
$(TARGET): $(OBJ)
	$(CC) mod tidy
	cd cmd/$(TARGET_NAME) && \
	$(CC) build -o ../../dist/$(TARGET_NAME)

# phony rules
.PHONY: all
all: $(TARGET)

.PHONY: clean
clean:
	@echo CLEAN $(CLEAN_LIST)
	@rm -f $(CLEAN_LIST)

.PHONY: distclean
distclean:
	@echo CLEAN $(CLEAN_LIST)
	@rm -f $(DISTCLEAN_LIST)

.PHONY: install
install:
	@cp -f dist/google-cloudflare-sync /usr/local/bin/

VERSION := $(shell git describe --tags --abbrev=0 | sed -Ee 's/^v|-.*//')
.PHONY: version
version:
	@echo v$(VERSION)

SEMVER_TYPES := major minor patch
BUMP_TARGETS := $(addprefix bump-,$(SEMVER_TYPES))
.PHONY: $(BUMP_TARGETS)
$(BUMP_TARGETS):
	$(eval bump_type := $(strip $(word 2,$(subst -, ,$@))))
	$(eval f := $(words $(shell a="$(SEMVER_TYPES)";echo $${a/$(bump_type)*/$(bump_type)} )))
	$(eval ver := $(shell echo $(VERSION) | awk -F. -v OFS=. -v f=$(f) '{ $$f++ } 1'))
	@echo $(ver)
	@sed -i "s/AppVersion =.*/AppVersion = \"$(ver)\"/g" cmd/google-cloudflare-sync/main.go
