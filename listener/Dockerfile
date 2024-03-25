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

# Build the application.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ../listener main.go

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/
FROM debian:buster-slim
RUN set -eux; apt-get update; apt-get install -y --no-install-recommends ca-certificates netbase

WORKDIR /

# Copy the pre-built binary file from the previous stage.
COPY --from=builder /app/listener /listener

# Command to run the executable
CMD ["/listener"]