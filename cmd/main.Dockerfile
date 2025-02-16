FROM golang:1.23.1-alpine AS builder
COPY . /github.com/Merch_store-Avito_test_task
WORKDIR /github.com/Merch_store-Avito_test_task
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o ./.bin ./cmd/main.go
FROM scratch AS runner
WORKDIR /build
COPY --from=builder /github.com/Merch_store-Avito_test_task/.bin .
EXPOSE 8080
ENTRYPOINT ["./.bin"]
