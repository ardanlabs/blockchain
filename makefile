SHELL := /bin/bash

# curl -il -X GET http://localhost:8080/v1/genesis/list
# curl -il -X GET http://localhost:8080/v1/balances/list
# curl -il -X GET http://localhost:8080/v1/blocks/list
# curl -X POST 'http:/localhost:8080/v1/tx/add' --header 'Content-Type: application/json' --data-raw '{"from": "bill_kennedy","to": "bill_kennedy","value": 10, "data": "reward"}'
# curl -X PUT 'http:/localhost:8080/v1/tx/persist'

# ==============================================================================
# Local support

service:
	go run app/services/barledger/main.go | go run app/tooling/logfmt/main.go

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