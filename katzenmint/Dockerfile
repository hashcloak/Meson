# source: katzenpost/server/Dockerfile
FROM golang:alpine as builder

LABEL authors="Peter Lai: peter@hashcloak.com"

# Install git & make
# Git is required for fetching the dependencies
RUN apk update && \
    apk add --no-cache git make ca-certificates build-base && \
    update-ca-certificates

# Set the working directory for the container
WORKDIR /go/katzenmint-pki

# Build the binary
COPY . .
RUN make build

FROM alpine

RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl && \
    update-ca-certificates

COPY --from=builder /go/katzenmint-pki/katzenmint /go/bin/katzenmint

# Expose the application port
# EXPOSE 8181

# create a volume for the configuration persistence
VOLUME /chain

# This form of ENTRYPOINT allows the process to catch signals from the `docker stop` command
ENTRYPOINT /go/bin/katzenmint -config /chain/katzenmint.toml
