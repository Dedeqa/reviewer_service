# Build stage
FROM golang:1.24.4 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o reviewer-service ./cmd/server

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/reviewer-service .
EXPOSE 8080
ENTRYPOINT ["./reviewer-service"]
