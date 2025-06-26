FROM golang:1.24-alpine3.22 AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

FROM alpine:3.22
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy static assets and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /root/data

EXPOSE 8080

CMD ["./main"]
