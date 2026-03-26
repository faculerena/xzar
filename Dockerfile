FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /xzar -ldflags="-s -w" .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && mkdir -p /data/uploads /etc/xzar
COPY --from=builder /xzar /usr/local/bin/xzar
VOLUME /data
ENV XZAR_ADDR=":8080" \
    XZAR_DOMAIN="xz.ar" \
    XZAR_DATA_DIR="/data" \
    XZAR_CREDENTIALS_FILE="/etc/xzar/credentials.json"
EXPOSE 8080
ENTRYPOINT ["xzar"]
