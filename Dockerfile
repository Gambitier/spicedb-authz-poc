FROM golang:1.23.3-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY default.yaml .

EXPOSE 8085 8086 9093
CMD ["./main"] 