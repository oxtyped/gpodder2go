create-migration:
	migrate create -seq -ext sql -dir . ${FILENAME}
.PHONY: create-migration

migrate-up:
	migrate -path=cmd/migrations/ -database sqlite3://${DB} up
.PHONY: migrate-up
migrate-down:
	migrate -path=cmd/migrations -database sqlite3://${DB} down
.PHONY: migrate-down

migrate-up-docker:
	docker run -v cmd/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database sqlite3://${DB} up 0
.PHONY: migrate-up-docker

build:
	go build -o gpodder2go main.go
.PHONY: build

# https://github.com/mvdan/gofumpt
# https://pkg.go.dev/golang.org/x/tools/cmd/goimports
fmt:
	go mod tidy
	gofumpt -l -w .
	goimports -w .
.PHONY: fmt

# https://golangci-lint.run/
lint:
	go vet ./...
	golangci-lint run --disable errcheck --enable sqlclosecheck --enable misspell --enable gofmt --enable goimports
.PHONY: lint

# https://go.dev/blog/vuln
vulncheck:
	govulncheck ./...
.PHONY: vulncheck
