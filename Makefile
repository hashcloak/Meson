all: build-docker
build-docker:
	docker build --no-cache -f Dockerfile -t hashcloack/meson ./src
