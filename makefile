SHELL := /bin/bash

# ==============================================================================
# Local support

run:
	go run app/services/barledger/main.go

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor