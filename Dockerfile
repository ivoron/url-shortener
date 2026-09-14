FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,https://goproxy.cn,direct

COPY go.mod go.sum ./

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /shortener ./cmd/shortener

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /shortener .

EXPOSE 8080

CMD ["./shortener"]
