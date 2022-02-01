SHELL := /bin/bash

# Seed transactions
# 
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-binary "@app/tooling/curl/seed.json" 
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-binary "@app/tooling/curl/trans1.json" 
#
# Add User transactions
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "bill_kennedy", "to": "bill_kennedy", "value": 10, "tip": 0, "data": "reward"}]'
# curl -X POST http:/localhost:8080/v1/tx/add --header 'Content-Type: application/json' --data-raw '[{"from": "ceasar", "to": "babayaga", "tip": 10, "value": 10}]'
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
# Wallet Stuff
# go run app/wallet/main.go generate - will generate new private key and store it in the private.ecdsa
# go run app/wallet/main.go balance - will pick private key from private.ecdsa and print out your balance
# go run app/wallet/main.go send -t bill_kennedy -v 15 - will send transaction to bill_kennedy with value 15

# ==============================================================================
# Local support

up:
	go run app/services/node/main.go | go run app/tooling/logfmt/main.go

up-race:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	go run app/services/node/main.go --web-debug-host 0.0.0.0:7181 --web-public-host 0.0.0.0:8180 --web-private-host 0.0.0.0:9180 --node-miner-account=miner2 --node-db-path zblock/blocks2.db | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep go-build | grep -v grep | sed -n 2,2p | cut -c1-5)

# ==============================================================================
# Wallet support

walbal:
	go run app/wallet/main.go balance

walgen:
	go run app/wallet/main.go generate

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

# ==============================================================================
# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck -checks=all ./...