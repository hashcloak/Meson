all: build-docker
build-docker:
	docker build --no-cache -t hashcloack/meson .
