# boot-dot-dev-build-a-blog-aggregator-in-go

- `go install github.com/pressly/goose/v3/cmd/goose@latest`
- `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- `go get github.com/google/uuid`
- `go get github.com/lib/pq`
- Connection string `postgres://postgres:postgres@localhost:5432/gator?sslmode=disable`
- On `./sql/schema` and run `goose postgres postgres://postgres:postgres@localhost:5432/gator up;`
