FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o api_server ./cmd/api_server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/api_server .
RUN chmod +x ./api_server && ls -la /app/api_server
EXPOSE 8080
CMD ["./api_server"]