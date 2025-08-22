GOBIN=${GOPATH}/bin

## Test:
test: ## Execute unit tests
	${GOBIN}/godotenv go test -race ./...

## Build:
build: ## Build project
	go build -o out/roistot cmd/server/server.go
	go build -o out/importer cmd/importer/importer.go

## Run:
run: ## Run project
	go run cmd/server/server.go

## Tidy:
tidy: ## Sync go.mod and go.sum with current status
	go mod tidy

## Update:
update: ## Update dependencies
	go get -t -u ./...
