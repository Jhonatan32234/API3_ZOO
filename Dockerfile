# Build stage
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api .

# Final stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/api .
RUN apk add --no-cache ca-certificates && adduser -D appuser
USER appuser
EXPOSE 8082
CMD ["./api"]
