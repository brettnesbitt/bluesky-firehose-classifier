# Multi-stage build to minimize image size

# Stage 1: Build the Go binary
FROM golang:1.22-alpine

# Install make
RUN apk add --no-cache make

# Install golangci-lint
RUN apk add --no-cache golangci-lint

WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy our files
COPY . .

COPY .env.docker .env

# Build the Go binary
RUN make build

RUN chmod +x /app/bin/bluesky-firehose-classifier

# Set the entrypoint
ENTRYPOINT ["/app/bin/bluesky-firehose-classifier"]

#TODO: Make Distroless

# Stage 2: Create the distroless image
#FROM gcr.io/distroless/static-debian11

#WORKDIR /

# Copy the binary from the builder stage
#COPY --from=builder /app/bin/bluesky-firehose-classifier /app/bluesky-firehose-classifier

# Set non root user
#USER 1000

# Set the entrypoint
#ENTRYPOINT ["/app/bluesky-firehose-classifier"]
