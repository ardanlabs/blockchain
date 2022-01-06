SHELL := /bin/bash

/*
	curl -il http://localhost:8080/v1/balances/list
	curl -il http://localhost:8080/v1/blocks/list

# ==============================================================================
# Local support

build:
	go build app/services/barledger/main.go

service:
	go run app/services/barledger/main.go | go run app/tooling/logfmt/main.go

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