GOBIN=${GOPATH}/bin

## Test:
test: ## Execute unit tests
	${GOBIN}/godotenv go test -race ./...

## Build:
build: ## Build project
	go build -o out/roistot server.go

## Run:
run: ## Run project
	go run server.go