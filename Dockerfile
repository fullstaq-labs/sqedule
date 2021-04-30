# Start from the latest golang base image
FROM golang:1.16.3 as builder

# Add Maintainer Info
LABEL maintainer="Fullstaq"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod, sum and main files
COPY . ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o sqedule -ldflags '-w -s' -a -installsuffix cgo ./cmd/sqedule-server

######## Start a new stage from scratch #######
FROM alpine:latest

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/sqedule .

# Expose port 3001 (not required)
EXPOSE 3001

# Using entrypoint so we can use commands in docker compose. Rather than using CMD
ENTRYPOINT ["./sqedule"] 