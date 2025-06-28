FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o code-refactoring-tool ./main.go

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/code-refactoring-tool .

EXPOSE 8080

CMD ["./code-refactoring-tool"]
