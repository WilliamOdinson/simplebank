FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN sqlc generate
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080
CMD ["/app/main"]
