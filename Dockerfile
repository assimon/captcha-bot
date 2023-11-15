FROM golang:latest as builder

WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y musl-tools

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    CC=x86_64-linux-musl-gcc \
    CXX=x86_64-linux-musl-g++ \
    GOPROXY=https://goproxy.cn,direct

RUN go mod download
RUN go mod verify
RUN go build -ldflags '-linkmode external -extldflags "-static"' -o captcha-bot

FROM alpine:latest

COPY --from=builder /app/captcha-bot /work/captcha-bot

WORKDIR /work

ENTRYPOINT ["./captcha-bot"]
