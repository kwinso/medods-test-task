set dotenv-load := true

migrate-down:
  migrate -database ${DATABASE_URL} -path sql/migrations down

gen-db:
  sqlc generate

serve:
  go run cmd/auth-server/main.go

dev:
  gow run cmd/auth-server/main.go

swagger:
  swag init --dir ./internal -g app.go