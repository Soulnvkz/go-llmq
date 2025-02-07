# Use the official Go image as a base
FROM golang:1.23.5-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o server .

# Use a lightweight final image
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the server runs on (change if needed)
EXPOSE 8080

# Run the Go server
CMD ["./server"]
