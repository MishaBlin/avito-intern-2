FROM golang:1.24 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o pvz-app ./cmd/pvz-app

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
RUN apk update && apk add --no-cache bash postgresql-client

WORKDIR /app

COPY --from=builder /src/pvz-app .
COPY ./internal/database/migrations /app/migrations
COPY scripts /app/scripts
RUN chmod +x /app/scripts/run-migrations.sh

CMD ["/app/pvz-app"]