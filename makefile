SHELL := /bin/bash

# curl -il -X GET http://localhost:8080/v1/genesis/list
# curl -il -X GET http://localhost:8080/v1/node/status
# curl -il -X GET http://localhost:8080/v1/balances/list
# curl -il -X GET http://localhost:8080/v1/blocks/list
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '{"from": "bill_kennedy","to": "bill_kennedy","value": 10, "data": "reward"}'
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '{"from": "ceasar","to": "babayaga","value": 10}'
# curl -X PUT http:/localhost:8080/v1/blocks/write

# ==============================================================================
# Local support

up1:
	go run app/services/barledger/main.go | go run app/tooling/logfmt/main.go

up2:
	go run app/services/barledger/main.go --web-api-host 0.0.0.0:8180 --web-debug-host 0.0.0.0:8181 --node-db-path zblock/blocks2.db | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep go-build | grep -v grep | sed -n 2,2p | cut -c1-5)

seed:
	go run app/tooling/admin/main.go trans seed

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