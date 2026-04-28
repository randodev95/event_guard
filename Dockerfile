# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o eventcanvas main.go

# Production Stage
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/eventcanvas .
COPY canvas.yaml .

# Standard ports
EXPOSE 8080

ENTRYPOINT ["./eventcanvas"]
CMD ["serve", "--plan", "canvas.yaml", "--port", "8080"]
