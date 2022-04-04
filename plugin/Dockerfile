ARG server=master
FROM golang:alpine AS builder
RUN apk update && \
  apk add --no-cache git make ca-certificates && \
  update-ca-certificates
WORKDIR /go/Meson
COPY . .
RUN go build -o Meson cmd/main.go

FROM hashcloak/server:${server}
COPY --from=builder /go/Meson/Meson /go/bin/Meson
ENTRYPOINT /go/bin/server -f /conf/katzenpost.toml
