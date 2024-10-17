FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-status-server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /user-status-server .

EXPOSE 8080

CMD ["./user-status-server"]