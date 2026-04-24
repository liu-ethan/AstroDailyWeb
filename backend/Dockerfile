FROM golang:1.25-alpine AS builder
WORKDIR /app

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/config /app/config

ENV CONFIG_PATH=/app/config/config.yaml

EXPOSE 9090
CMD ["/app/server"]
