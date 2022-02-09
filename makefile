SHELL := /bin/bash

# Wallets
# Kennedy: 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32
# Pavel: 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4
# Ceasar: 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76
# Baba: 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9
#
# Bookeeping transactions
# curl -il -X GET http://localhost:8080/v1/genesis/list
# curl -il -X GET http://localhost:9080/v1/node/status
# curl -il -X GET http://localhost:8080/v1/balances/list
# curl -il -X GET http://localhost:8080/v1/tx/uncommitted/list
# curl -il -X GET http://localhost:8080/v1/blocks/list
# curl -il -X GET http://localhost:8080/v1/blocks/list/1/latest
#
# Force a mining operation
# curl -il -X GET http://localhost:8080/v1/mining/signal
#
# Wallet Stuff
# go run app/wallet/main.go generate
# go run app/wallet/main.go account -a kennedy
# go run app/wallet/main.go balance -a kennedy
# go run app/wallet/main.go send -a kennedy -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
# go run app/wallet/main.go send -a kennedy -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 150
# go run app/wallet/main.go send -a pavel -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
# go run app/wallet/main.go send -a pavel -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 250

# ==============================================================================
# Local support

up:
	CGO_ENABLED=0 go run app/services/node/main.go | go run app/tooling/logfmt/main.go

up-race:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	CGO_ENABLED=0 go run app/services/node/main.go --web-debug-host 0.0.0.0:7181 --web-public-host 0.0.0.0:8180 --web-private-host 0.0.0.0:9180 --node-miner-address=miner2 --node-db-path zblock/blocks2.db | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep go-build | grep -v grep | sed -n 2,2p | cut -c1-5)

# ==============================================================================
# Wallet support

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