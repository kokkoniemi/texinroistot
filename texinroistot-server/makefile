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
