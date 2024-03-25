# Start from the latest golang base image
FROM golang:1.22.0-alpine3.19 as builder

# Set the Current Working Directory inside the container.
WORKDIR /app

# Copy go module files and download dependencies.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the source code into the container.
COPY src/ ./src/

# Set the Current Working Directory inside the container to the src directory.
WORKDIR /app/src

# Build the application, outputting the executable to /bin directory.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /bin/listener main.go

# Use a Docker multi-stage build to create a lean production image.
FROM debian:bookworm-slim
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates netbase; \
    # Remove apt cache to reduce image size.
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /

# Copy the pre-built binary file from the previous stage.
COPY --from=builder /bin/listener /listener

# Command to run the executable
CMD ["/listener"]
