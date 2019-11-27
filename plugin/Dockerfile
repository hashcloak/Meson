FROM golang:alpine AS builder

# Install git & make
# Git is required for fetching the dependencies
RUN apk update && \
    apk add --no-cache git make ca-certificates && \
    update-ca-certificates

# Set the working directory for the container
WORKDIR /go/Meson

# Build the binary
COPY . .
RUN go build -o Meson cmd/main.go 

FROM katzenpost/server

COPY --from=builder /go/Meson/Meson /go/bin/Meson

# Expose the application port
# EXPOSE 8181

# create a volume for the configuration persistence
VOLUME /conf

# This form of ENTRYPOINT allows the process to catch signals from the `docker stop` command
ENTRYPOINT /go/bin/server -f /conf/katzenpost.toml
