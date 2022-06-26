FROM golang:alpine as builder

LABEL authors="Peter Lai: peter@hashcloak.com"

RUN apk update && \
    apk add --no-cache git make ca-certificates build-base && \
    update-ca-certificates

WORKDIR /go

RUN git clone https://github.com/hashcloak/Meson.git
RUN cd Meson/katzenmint ; make build ;

FROM alpine

RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl && \
    update-ca-certificates

COPY --from=builder /go/Meson/katzenmint/katzenmint /go/bin/katzenmint

# EXPOSE 8181

VOLUME /chain

ENTRYPOINT /go/bin/katzenmint run --config /chain/katzenmint.toml
