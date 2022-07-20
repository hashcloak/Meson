.PHONY: build-docker-katzenmint
build-docker-katzenmint:
	docker build --no-cache -t katzenmint/pki -f Dockerfile.katzenmint .

.PHONY: build-docker-server
build-docker-server:
	docker build --no-cache -t meson/server -f Dockerfile.server .

.PHONY: build-docker-containers
build-docker-containers: build-docker-katzenmint build-docker-server

.PHONY: clean-docker-images
clean-docker-images:
	docker rmi -f $$(docker images | grep '^<none>' | awk '{print $$3}')