SHELL := /bin/bash

# Wallets
# Kennedy: 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32
# Pavel: 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4
# Ceasar: 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76
# Baba: 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9
# Ed: 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0
# Miner1: 0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8
# Miner2: 0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61
#
# Run two miners
# make up
# make up2
#
# Bookeeping transactions
# curl -il -X GET http://localhost:8080/v1/genesis/list
# curl -il -X GET http://localhost:9080/v1/node/status
# curl -il -X GET http://localhost:8080/v1/accounts/list
# curl -il -X GET http://localhost:8080/v1/tx/uncommitted/list
# curl -il -X GET http://localhost:8080/v1/blocks/list
# curl -il -X GET http://localhost:9080/v1/node/block/list/1/latest
#
# Wallet Stuff
# go run app/wallet/cli/main.go generate
# go run app/wallet/cli/main.go account -a kennedy
# go run app/wallet/cli/main.go balance -a kennedy

# ==============================================================================
# Local support

up:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7281 --web-public-host 0.0.0.0:8280 --web-private-host 0.0.0.0:9280 --state-beneficiary=miner2 --state-db-path zblock/miner2/ | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep "main -race" | grep -v grep | sed -n 1,1p | cut -c1-5)

down-ubuntu:
	kill -INT $(shell ps -x | grep "main -race" | sed -n 1,1p | cut -c3-7)

# ==============================================================================
# Docker support

docker-up:
	docker compose -f zarf/docker/docker-compose.yml up

docker-down:
	docker compose -f zarf/docker/docker-compose.yml down

docker-logs:
	docker compose -f zarf/docker/docker-compose.yml logs

# ==============================================================================
# Transactions

load:
	go run app/wallet/cli/main.go send -a kennedy -n 1 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/wallet/cli/main.go send -a pavel -n 1 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
	go run app/wallet/cli/main.go send -a kennedy -n 2 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 150
	go run app/wallet/cli/main.go send -a pavel -n 2 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 125
	go run app/wallet/cli/main.go send -a kennedy -n 3 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 200
	go run app/wallet/cli/main.go send -a pavel -n 3 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 250

load2:
	go run app/wallet/cli/main.go send -a kennedy -n 4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/wallet/cli/main.go send -a pavel -n 4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75

load3:
	go run app/wallet/cli/main.go send -a kennedy -n 5 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 150
	go run app/wallet/cli/main.go send -a pavel -n 5 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 125
	go run app/wallet/cli/main.go send -a kennedy -n 6 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 200
	go run app/wallet/cli/main.go send -a pavel -n 6 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 250

# ==============================================================================
# Viewer support

react:
	npm install --prefix app/services/viewer/
	npm start --prefix app/services/viewer/

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

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -v ./...
	go mod tidy
	go mod vendor

# ==============================================================================
# Running tests within the local computer
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	go test ./... -count=1
	staticcheck -checks=all ./...
	govulncheck ./...