FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies and create go.sum
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o soltar-vpn ./cmd/worker

# Copy webapp files
COPY webapp/ /app/webapp/

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/soltar-vpn .

# Copy webapp files
COPY --from=builder /app/webapp ./webapp

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./soltar-vpn"] 