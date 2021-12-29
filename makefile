SHELL := /bin/bash

# ==============================================================================
# Local support

service:
	go run app/services/barledger/main.go

bals:
	go run app/tooling/admin/main.go bals

trans:
	go run app/tooling/admin/main.go trans

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor