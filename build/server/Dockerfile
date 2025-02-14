# Use the official Go image as a base
FROM golang:1.23.5-alpine AS builder

# Set the working directory inside the container
WORKDIR /project

COPY ./server ./server
COPY ./pkg ./pkg

# Download dependencies
RUN go mod tidy -C ./server

# Build the Go application
RUN go build -C ./server/cmd -o server

# Use a lightweight final image
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /project/server/cmd/server .

# Expose the port the server runs on (change if needed)
EXPOSE 8080

# Run the Go server
CMD ["./server"]
