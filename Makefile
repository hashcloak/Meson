flags=.makeFlags
VPATH=$(flags)
$(shell mkdir -p $(flags))
dockerRepo=hashcloak
katzenBranch=master
katzenServer=$(dockerRepo)/katzenpost-server:$(katzenBranch)
katzenAuth=$(dockerRepo)/katzenpost-auth:$(katzenBranch)
gethVersion=v1.9.9
gethImage=$(dockerRepo)/client-go:$(gethVersion)
mesonServer=$(dockerRepo)/meson
mesonClient=$(dockerRepo)/meson-client

GIT_HASH := $(shell git log --format='%h' -n1)
TRAVIS_BRANCH ?= $(shell git log --format='%D' -n1 | cut -d'/' -f2)
BRANCH=$(TRAVIS_BRANCH)

.PHONY: up down

clean:
	rm -rf /tmp/server
	rm -rf /tmp/authority
	rm -rf /tmp/Meson
	rm -rf $(flags)

clean-data:
	rm -r $(flags)/permits
	rm -rf ./ops/nonvoting_testnet/conf/provider?
	rm -rf ./ops/nonvoting_testnet/conf/mix?
	rm -rf ./ops/nonvoting_testnet/conf/auth
	git checkout ./ops/nonvoting_testnet/conf
	$(MAKE) permits

pull: pull-tags

pull-tags: pull-katzen-server
	docker pull $(katzenAuth) && echo "success!" || $(MAKE) katzenpost-nonvoting-authority
	docker pull $(gethImage) && echo "success!" || $(MAKE) hashcloak-geth

pull-katzen-server:
	docker pull $(katzenServer) && echo "success!" || $(MAKE) katzenpost-server

push-tags: build-images
	docker push $(katzenServer)
	docker push $(katzenAuth)
	docker push '$(mesonServer):$(BRANCH)'
	docker push $(gethImage)

build-images: hashcloak-geth katzenpost-server katzenpost-nonvoting-authority meson

hashcloak-geth:
	sed -i 's|%%GETH_VERSION%%|$(gethVersion)|g' ./ops/geth.Dockerfile
	docker build -f ./ops/geth.Dockerfile -t $(gethImage) .
	sed -i 's|$(gethVersion)|%%GETH_VERSION%%|g' ./ops/geth.Dockerfile
	@touch $(flags)/$@

katzenpost-server:
	git clone https://github.com/katzenpost/server /tmp/server || true
	docker build -f /tmp/server/Dockerfile -t $(katzenServer) /tmp/server
	@touch $(flags)/$@

katzenpost-nonvoting-authority:
	git clone https://github.com/katzenpost/authority /tmp/authority || true
	docker build -f /tmp/authority/Dockerfile.nonvoting -t $(katzenAuth) /tmp/authority
	@touch $(flags)/$@

meson: pull-katzen-server hashcloak-geth
	sed -i 's|%%KATZEN_SERVER%%|$(katzenServer)|g' ./plugin/Dockerfile
	docker build -f ./plugin/Dockerfile -t $(mesonServer):$(BRANCH) ./plugin
	sed -i 's|$(katzenServer)|%%KATZEN_SERVER%%|g' ./plugin/Dockerfile
	@touch $(flags)/$@

up: permits pull meson up-nonvoting

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

rebuild-meson:
	docker build -f ./Dockerfile -t $(mesonServer):latest .

test-client:
	git clone https://github.com/hashcloak/Meson-client /tmp/Meson-client || true
	docker run \
		-v /tmp/Meson-client:/client \
		-v /tmp/gopath-pkg:/go/pkg \
		--network nonvoting_testnet_nonvoting_test_net \
		-w /client \
		golang:buster \
		/bin/bash -c "GORACE=history_size=7 go test -race"

failure:
	false && ([ $$? -eq 0 ] && echo "success!") || echo "failure"

success:
	true && ([ $$? -eq 0 ] && echo "success!") || echo "failure!"     
