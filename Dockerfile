FROM golang:1.24-alpine3.22 AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

FROM alpine:3.22
RUN apk --no-cache add ca-certificates sqlite

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Switch to app user
USER appuser
WORKDIR /home/appuser/

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/main .

# Copy static assets and templates
COPY --from=builder --chown=appuser:appgroup /app/migrations ./migrations

# Create data directory with proper permissions
RUN mkdir -p /home/appuser/data && chmod 700 /home/appuser/data

EXPOSE 8080

CMD ["./main"]
