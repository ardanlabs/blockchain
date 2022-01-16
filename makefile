SHELL := /bin/bash

# Seed transactions
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "bill_kennedy", "to": "bill_kennedy", "value": 3, "data": "reward"},{"from": "bill_kennedy", "to": "bill_kennedy", "value": 703, "data": "reward"}]'
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "bill_kennedy", "to": "babayaga", "value": 2000, "data": ""},{"from": "bill_kennedy", "to": "bill_kennedy", "value": 100, "data": "reward"},{"from": "babayaga", "to": "bill_kennedy", "value": 1, "data": ""},{"from": "babayaga", "to": "ceasar", "value": 1000, "data": ""},{"from": "babayaga", "to": "bill_kennedy", "value": 50, "data": ""},{"from": "bill_kennedy", "to": "bill_kennedy", "value": 600, "data": "reward"}]'
#
# Extra transactions
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "bill_kennedy","to": "bill_kennedy","value": 10, "data": "reward"}]'
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "ceasar","to": "babayaga","value": 10}]'
#
# Bookeeping transactions
# curl -il -X GET http://localhost:8080/v1/genesis/list
# curl -il -X GET http://localhost:8080/v1/node/status
# curl -il -X GET http://localhost:8080/v1/balances/list
# curl -il -X GET http://localhost:8080/v1/blocks/list
# curl -il -X GET http://localhost:8080/v1/blocks/list/1/latest
#
# Force a mining operation
# curl -il -X GET http://localhost:8080/v1/mining/signal
#

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