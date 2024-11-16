# Stage 1: Build the application
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application code
COPY . .

# Build the Go application with the architecture set to amd64 (if needed)
RUN GOARCH=amd64 go build -o /app/main ./app/main.go

# Stage 2: Create a minimal image for running the app
FROM alpine:latest

# Install certificates for HTTPS connections (optional)
RUN apk add --no-cache ca-certificates

# Set the working directory
WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Make sure the binary is executable
RUN chmod +x /app/main

# Copy the .env file (optional)
COPY .env .env

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]
