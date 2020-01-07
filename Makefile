TRAVIS_BRANCH ?= $(shell git branch| grep \* | cut -d' ' -f2)
BRANCH=$(TRAVIS_BRANCH)

ifdef $(TRAVIS_PULL_REQUEST_BRANCH)
ifneq ($(TRAVIS_PULL_REQUEST_BRANCH),$(TRAVIS_BRANCH))
	BRANCH = $(TRAVIS_PULL_REQUEST_BRANCH)
endif
endif

flags=.makeFlags
VPATH=$(flags)
$(shell mkdir -p $(flags))
dockerRepo=hashcloak

katzenServerRepo=https://github.com/katzenpost/server
katzenServerTag=$(shell git ls-remote --heads $(katzenServerRepo) | grep master | cut -c1-7)
katzenServer=$(dockerRepo)/katzenpost-server:$(katzenServerTag)

katzenAuthRepo=https://github.com/katzenpost/authority
katzenAuthTag=$(shell git ls-remote --heads https://github.com/katzenpost/authority  | grep master | cut -c1-7)
katzenAuth=$(dockerRepo)/katzenpost-auth:$(katzenAuthTag)

gethVersion=v1.9.9
gethImage=$(dockerRepo)/client-go:$(gethVersion)
mesonServer=$(dockerRepo)/meson
mesonClient=$(dockerRepo)/meson-client

messagePush=echo "LOG: Image already exists in docker.io/$(dockerRepo). Not pushing: "
messagePull=echo "LOG: Success in pulling image: "
imageNotFound=echo "LOG: Image not found... building: "

clean:
	rm -rf /tmp/server
	rm -rf /tmp/authority
	rm -rf $(flags)

clean-data:
	rm -rf ./ops/nonvoting_testnet/conf/provider?
	rm -rf ./ops/nonvoting_testnet/conf/mix?
	rm -rf ./ops/nonvoting_testnet/conf/auth
	git checkout ./ops/nonvoting_testnet/conf
	rm -r $(flags)/permits || true
	$(MAKE) permits

pull: pull-katzen-auth pull-katzen-server pull-geth

pull-katzen-auth:
	docker pull $(katzenAuth) && $(messagePull)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && $(MAKE) build-katzen-nonvoting-authority)

pull-katzen-server:
	docker pull $(katzenServer) && $(messagePull)$(katzenServer) \
		|| ($(imageNotFound)$(katzenServer) && $(MAKE) build-katzen-server)

pull-geth:
	docker pull $(gethImage) && $(messagePull)$(gethImage) \
		||  ($(imageNotFound)$(gethImage) && $(MAKE) build-geth)

push: push-katzen-server push-katzen-auth push-geth push-meson

push-katzen-server:
	docker push $(katzenServer) && $(messagePush)$(katzenServer) \
		|| ($(imageNotFound)$(katzenServer) && \
				$(MAKE) build-katzen-server && \
				docker push $(katzenServer))

push-katzen-auth:
	docker push $(katzenAuth) && $(messagePush)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && \
				$(MAKE) build-katzen-nonvoting-authority && \
				docker push $(katzenAuth))

push-geth:
	docker push $(gethImage) && $(messagePush)$(gethImage) \
		|| ($(imageNotFound)$(gethImage) && \
				$(MAKE) build-geth && \
				docker push $(gethImage))

push-meson: build-meson
	docker push '$(mesonServer):$(BRANCH)'

build: build-geth build-katzen-server build-katzen-nonvoting-authority build-meson

build-geth:
	sed 's|%%GETH_VERSION%%|$(gethVersion)|g' ./ops/geth.Dockerfile > /tmp/geth.Dockerfile
	docker build -f /tmp/geth.Dockerfile -t $(gethImage) .
	@touch $(flags)/$@

build-katzen-server:
	git clone $(katzenServerRepo) /tmp/server || true
	docker build -f /tmp/server/Dockerfile -t $(katzenServer) /tmp/server
	@touch $(flags)/$@

build-katzen-nonvoting-authority:
	git clone $(katzenAuthRepo) /tmp/authority || true
	docker build -f /tmp/authority/Dockerfile.nonvoting -t $(katzenAuth) /tmp/authority
	@touch $(flags)/$@

build-meson: pull-katzen-server
	sed 's|%%KATZEN_SERVER%%|$(katzenServer)|g' ./plugin/Dockerfile > /tmp/meson.Dockerfile
	docker build -f /tmp/meson.Dockerfile -t $(mesonServer):$(BRANCH) ./plugin
	@touch $(flags)/$@

up: permits pull build-meson up-nonvoting

permits:
	sudo chmod -R 700 ops/nonvoting_testnet/conf/provider?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/mix?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/auth
	@touch $(flags)/$@

up-nonvoting:
	GETH_IMAGE=$(gethImage) \
	KATZEN_SERVER=$(katzenServer) \
	KATZEN_AUTH=$(katzenAuth) \
	MESON_IMAGE=$(mesonServer):$(BRANCH) \
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml up -d

down:
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml down

test-client:
	git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client || true
	docker run \
		-v /tmp/Meson-client:/client \
		-v /tmp/gopath-pkg:/go/pkg \
		--network nonvoting_testnet_nonvoting_test_net \
		-w /client \
		golang:buster \
		/bin/bash -c "GORACE=history_size=7 go test -race"
	[ ${CI} ]; sudo chown ${USER} -R /tmp/gopath-pkg
