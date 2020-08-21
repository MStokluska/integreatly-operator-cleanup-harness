DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
OUT_FILE := "$(DIR)integreatly-operator-cleanup-harness"

build:
	go mod vendor
	CGO_ENABLED=0 go test -v -c
