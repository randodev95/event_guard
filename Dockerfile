# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
# Install build dependencies
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o event_guard main.go

# Production Stage
FROM alpine:3.21

# Security: Run as non-privileged user
RUN adduser -D -u 10001 eventguard
USER eventguard

WORKDIR /home/eventguard/
COPY --from=builder /app/event_guard .
COPY maps/ maps/

# Default Environment Variables
ENV EVENT_GUARD_PORT=8080
ENV EVENT_GUARD_PLAN=maps/

# Standard ports
EXPOSE 8080

ENTRYPOINT ["./event_guard"]
CMD ["serve", "--plan", "maps/", "--port", "8080"]
