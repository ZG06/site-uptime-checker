# --- Stage 1: Build Stage ---
FROM golang:1.27-alpine AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code.
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o site-checker main.go


# --- Stage 2: Minimal Runtime Stage ---
FROM alpine:latest

# Install root CA certificates for outgoing HTTP / TLS calls (fetching data from AWS S3)
RUN apk --no-cache add ca-certificates

# Set destination for COPY
WORKDIR /app

# Copy the compiled binary from Stage 1 into system PATH
COPY --from=builder /app/site-checker /usr/local/bin/site-checker

# Set the default entrypoint command to run the binary
ENTRYPOINT ["site-checker"]