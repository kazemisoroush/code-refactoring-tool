FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o code-refactor-tool ./main.go

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/code-refactor-tool .

EXPOSE 8080

CMD ["./code-refactor-tool"]
