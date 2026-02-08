FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o sendrec .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and static files from builder
COPY --from=builder /app/sendrec .
COPY --from=builder /app/static ./static

# Create data directory for waitlist storage
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Run
CMD ["./sendrec"]
