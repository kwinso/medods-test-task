set dotenv-load := true

migrate-up:
  migrate -database ${DB_URL} -path sql/migrations up

migrate-down:
  migrate -database ${DB_URL} -path sql/migrations down

gen-db:
  sqlc generate

serve:
  go run cmd/auth-server/main.go

dev:
  gow run cmd/auth-server/main.go

swagger:
  swag init --dir ./internal -g app.go