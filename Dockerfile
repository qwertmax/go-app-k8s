FROM golang:latest
LABEL maintainer="qwertmax@gmail.com"

WORKDIR /app

COPY main.go .

RUN go build -o app .

EXPOSE 80
