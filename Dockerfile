FROM golang:latest
LABEL maintainer="qwertmax@gmail.com"

WORKDIR /app
COPY main.go .
RUN go build -o app .
ARG VERSION
ENV VERSION=$VERSION
CMD ["./app"]

EXPOSE 80