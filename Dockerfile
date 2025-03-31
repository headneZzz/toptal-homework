FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bookshop ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bookshop .

COPY migrations ./migrations

EXPOSE 8080

CMD ["./bookshop"]