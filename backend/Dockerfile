FROM golang:1.25-alpine AS builder

WORKDIR /src

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /out/server /app/server
COPY config /app/config
COPY internal/notify /app/internal/notify
COPY internal/llm/prompt_template.yaml /app/internal/llm/prompt_template.yaml

EXPOSE 9090

ENTRYPOINT ["/app/server"]
