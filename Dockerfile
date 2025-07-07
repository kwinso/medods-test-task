FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth_server ./cmd/auth_server/main.go

FROM gcr.io/distroless/static-debian12 AS runner

WORKDIR /

COPY --from=build-stage /auth_server /auth_server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/auth_server"]
