# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o event_guard main.go

# Production Stage
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/event_guard .
COPY canvas.yaml .

# Standard ports
EXPOSE 8080

ENTRYPOINT ["./event_guard"]
CMD ["serve", "--plan", "canvas.yaml", "--port", "8080"]
