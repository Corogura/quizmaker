# Build stage
FROM --platform=linux/amd64 golang:latest AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o quizmaker

# Final stage
FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/quizmaker .
RUN chmod +x quizmaker

# Copy static files
COPY static ./static

CMD ["./quizmaker"]
