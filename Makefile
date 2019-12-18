TRAVIS_BRANCH ?= $(shell git branch| grep \* | cut -d' ' -f2)
BRANCH=$(TRAVIS_BRANCH)

ifdef $(TRAVIS_PULL_REQUEST_BRANCH)
ifneq ($(TRAVIS_PULL_REQUEST_BRANCH), "")
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
katzenAuthTag=$(shell git ls-remote --heads $(katzenAuthRepo) | grep master | cut -c1-7)
katzenAuth=$(dockerRepo)/katzenpost-auth:$(katzenAuthTag)

gethVersion=v1.9.9
gethImage=$(dockerRepo)/client-go:$(gethVersion)
mesonServer=$(dockerRepo)/meson
hashcloakAuth=$(dockerRepo)/authority

messagePush="LOG: Image already exists in docker.io/$(repoRepo). Not pushing: "
messagePull="LOG: success in pulling image: "

clean:
	rm -rf /tmp/server
	rm -rf /tmp/authority
	rm -rf $(flags)
	mkdir $(flags)

clean-data:
	rm -rf ./ops/nonvoting_testnet/conf/provider?
	rm -rf ./ops/nonvoting_testnet/conf/mix?
	rm -rf ./ops/nonvoting_testnet/conf/auth
	git checkout ./ops/nonvoting_testnet/conf
	rm -r $(flags)/permits || true
	$(MAKE) permits

pull:
	docker pull $(katzenAuth) \
		&& echo $(messagePull)$(katzenAuth) \
		|| $(MAKE) build-katzen-nonvoting-authority
	docker pull $(katzenServer) \
		&& echo $(messagePull)$(katzenServer) \
		|| $(MAKE) build-katzen-server
	docker pull $(gethImage) \
		&& echo $(messagePull)$(gethImage) \
		|| $(MAKE) build-geth
	@touch $(flags)/$@

push: push-katzen-server push-katzen-auth push-geth push-meson push-hashcloak-nonvoting-auth

push-katzen-server:
	docker pull $(katzenServer) \
		&& echo $(messagePush)$(katzenServer) \
		|| ($(MAKE) build-katzen-server && docker push $(katzenServer))

push-katzen-auth:
	docker pull $(katzenAuth) \
		&& echo $(messagePush)$(katzenAuth) \
		|| ($(MAKE) build-katzen-nonvoting-authority && docker push $(katzenAuth))

push-geth:
	docker pull $(gethImage) \
		&& echo $(messagePush)$(gethImage) \
		|| ($(MAKE) build-geth && docker push $(gethImage))

push-hashcloak-nonvoting-auth: build-hashcloak-nonvoting-authority
	docker push '$(hashcloakAuth):$(BRANCH)'

push-meson: build-meson
	docker push '$(mesonServer):$(BRANCH)'

build: build-geth build-katzen-server build-katzen-nonvoting-authority build-meson build-hashcloak-nonvoting-authority

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

build-meson: pull
	sed 's|%%KATZEN_SERVER%%|$(katzenServer)|g' ./ops/meson.Dockerfile > /tmp/meson.Dockerfile
	docker build -f /tmp/meson.Dockerfile -t $(mesonServer):$(BRANCH) .
	@touch $(flags)/$@

build-hashcloak-nonvoting-authority: pull
	sed 's|%%KATZENPOST_AUTH%%|$(katzenAuth)|g' ./ops/auth.nonvoting.Dockerfile > /tmp/auth.nonvoting.Dockerfile
	docker build -f /tmp/auth.nonvoting.Dockerfile -t $(hashcloakAuth):$(BRANCH) ./ops
	@touch $(flags)/$@

permits:
	sudo chmod -R 700 ops/nonvoting_testnet/conf/provider?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/mix?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/auth
	@touch $(flags)/$@

up: permits build-meson build-hashcloak-nonvoting-authority
	GETH_IMAGE=$(gethImage) \
	KATZEN_SERVER=$(katzenServer) \
	KATZEN_AUTH=$(katzenAuth) \
	MESON_IMAGE=$(mesonServer):$(BRANCH) \
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml up -d

up-dynamic: build-meson build-hashcloak-nonvoting-authority
	GETH_IMAGE=$(gethImage) \
	AUTHORITY_IMAGE=$(hashcloakAuth):$(BRANCH) \
	MESON_IMAGE=$(mesonServer):$(BRANCH) \
	docker-compose -f ./ops/dynamic-testnet.yml up -d

down:
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml down
	docker-compose -f ./ops/dynamic-testnet.yml down

test-client:
	git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client || true
	docker run \
		-v /tmp/Meson-client:/client \
		-v /tmp/gopath-pkg:/go/pkg \
		--network $(shell docker network ls --format '{{.Name}}' | grep testnet) \
		-w /client \
		golang:buster \
		/bin/bash -c "GORACE=history_size=7 go test -race"
	if [ ${CI} ];then sudo chown ${USER} -R /tmp/gopath-pkg;  fi
