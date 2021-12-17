SHELL := /bin/bash

# ==============================================================================
# Local support

service:
	go run app/services/barledger/main.go

balances:
	go run app/tooling/admin/main.go balances

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor