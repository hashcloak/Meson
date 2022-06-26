FROM golang:alpine AS builder

LABEL authors="Christian Muehlhaeuser: muesli@gmail.com"

# Can pass --build-arg warped=true to decrease epoch period
ARG warped=false

ENV ldflags="-X github.com/katzenpost/core/epochtime.WarpedEpoch=${warped} -X github.com/hashcloak/Meson/server/internal/pki.WarpedEpoch=${warped} -X github.com/katzenpost/minclient/pki.WarpedEpoch=${warped}"

RUN apk update && \
    apk add --no-cache git make ca-certificates && \
    update-ca-certificates

WORKDIR /go

RUN cd /go ; git clone https://github.com/hashcloak/Meson.git
RUN cd /go/Meson/server && go build -o meson-server -ldflags "$ldflags" cmd/meson-server/*.go
RUN cd /go ; git clone https://github.com/katzenpost/memspool.git
RUN cd /go ; cd memspool/server/cmd/memspool ;  go build -ldflags "$ldflags"
# RUN cd /go ; cd katzenpost/reunion ; cd servers/reunion_katzenpost_server ; go build -ldflags "$ldflags"
RUN cd /go ; git clone https://github.com/katzenpost/panda.git
RUN cd /go ; cd panda/server/cmd/panda_server ; go build -ldflags "$ldflags"
RUN cd /go ; git clone https://github.com/katzenpost/server_plugins.git
RUN cd /go ; cd server_plugins/cbor_plugins/echo-go ; go build -o echo_server -ldflags "$ldflags"
RUN cd /go ; cd Meson/plugin ; go build -o Meson -ldflags "$ldflags" cmd/meson-plugin/main.go
RUN cd /go ; git clone https://github.com/hashcloak/genconfig ; cd genconfig ; git pull origin add-cilint ; go build -o updateconfig -ldflags "$ldflags" update/main.go

FROM alpine

RUN apk update && \
    apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

COPY --from=builder /go/Meson/server/meson-server /go/bin/server
COPY --from=builder /go/memspool/server/cmd/memspool/memspool /go/bin/memspool
# COPY --from=builder /go/katzenpost/reunion/servers/reunion_katzenpost_server/reunion_katzenpost_server /go/bin/reunion_katzenpost_server
COPY --from=builder /go/panda/server/cmd/panda_server/panda_server /go/bin/panda_server
COPY --from=builder /go/server_plugins/cbor_plugins/echo-go/echo_server /go/bin/echo_server
COPY --from=builder /go/Meson/plugin/Meson /go/bin/Meson
COPY --from=builder /go/genconfig/updateconfig /go/bin/updateconfig

# EXPOSE 8181

VOLUME /conf

ENTRYPOINT /go/bin/server -f /conf/katzenpost.toml
