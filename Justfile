set dotenv-load := true

migrate-up:
  migrate -database ${AUTH_DB_URL} -path sql/migrations up

migrate-down:
  migrate -database ${AUTH_DB_URL} -path sql/migrations down -all

gen-db:
  sqlc generate

serve:
  go run cmd/auth_server/main.go

dev:
  gow run cmd/auth_server/main.go

swagger:
  swag init --dir ./internal -g app.go