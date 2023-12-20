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

.PHONE: test-all
test-all:
	@$(MAKE) $(MAKE_FLAGS) -C client/. test
	@$(MAKE) $(MAKE_FLAGS) -C genconfig/. test
	@$(MAKE) $(MAKE_FLAGS) -C katzenmint/. test
	@$(MAKE) $(MAKE_FLAGS) -C plugin/. test
	@$(MAKE) $(MAKE_FLAGS) -C server/. test

.PHONE: build-all
build-all:
	@$(MAKE) $(MAKE_FLAGS) -C genconfig/. build
	@$(MAKE) $(MAKE_FLAGS) -C katzenmint/. build
	@$(MAKE) $(MAKE_FLAGS) -C plugin/. build
	@$(MAKE) $(MAKE_FLAGS) -C server/. build

.PHONE: all
all:
	@$(MAKE) $(MAKE_FLAGS) -C client/.
	@$(MAKE) $(MAKE_FLAGS) -C genconfig/.
	@$(MAKE) $(MAKE_FLAGS) -C katzenmint/.
	@$(MAKE) $(MAKE_FLAGS) -C plugin/.
	@$(MAKE) $(MAKE_FLAGS) -C server/.
