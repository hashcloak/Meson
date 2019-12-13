flags=.makeFlags
VPATH=$(flags)
$(shell mkdir -p $(flags))
gethVersion=v1.9.8
dockerRepo=hashcloak
katzenServer=$(dockerRepo)/katzenpost-server
katzenAuth=$(dockerRepo)/katzenpost-auth
katzenBranch=master
gethImage=$(dockerRepo)/client-go:$(gethVersion)
mesonServer=$(dockerRepo)/meson
mesonClient=$(dockerRepo)/meson-client

GIT_HASH := $(shell git log --format='%h' -n1)
BRANCH := $(shell git log --format='%D' -n1 | cut -d'/' -f2)

.PHONY:up down

all: hashcloak-geth katzenpost-server meson katzenpost-nonvoting-authority

clean:
	rm -rf /tmp/server
	rm -rf /tmp/authority
	rm -rf /tmp/Meson
	rm -rf .makeFlags

clean-data:
	rm -rf ./ops/nonvoting_testnet/conf/provider?
	rm -rf ./ops/nonvoting_testnet/conf/mix?
	rm -rf ./ops/nonvoting_testnet/conf/auth
	git checkout ./ops/nonvoting_testnet/conf
	$(MAKE) permits


pull: pull-tags

pull-tags:
	docker pull '$(katzenServer):$(BRANCH)'
	docker pull '$(katzenAuth):$(BRANCH)'
	docker pull '$(mesonServer):$(BRANCH)'

push-tags: katzenpost-server katzenpost-nonvoting-authority meson
	docker push '$(katzenServer):$(katzenBranch)'
	docker push '$(katzenAuth):$(katzenBranch)'
	docker push '$(mesonServer):$(BRANCH)'
	docker push $(gethImage)

build-images: hashcloak-geth katzenpost-server katzenpost-nonvoting-authority meson

hashcloak-geth:
	sed -i 's|%%GETH_VERSION%%|$(gethVersion)|g' ./ops/geth.Dockerfile
	docker build -f ./ops/geth.Dockerfile -t $(dockerRepo)/$(gethImage):$(gethVersion) .
	sed -i 's|$(gethVersion)|%%GETH_VERSION%%|g' ./ops/geth.Dockerfile
	@touch $(flags)/$@

katzenpost-server:
	git clone https://github.com/katzenpost/server /tmp/server || true
	docker build -f /tmp/server/Dockerfile -t $(katzenServer):$(katzenBranch) /tmp/server
	@touch $(flags)/$@

katzenpost-nonvoting-authority:
	git clone https://github.com/katzenpost/authority /tmp/authority || true
	docker build -f /tmp/authority/Dockerfile.nonvoting -t $(katzenAuth):$(katzenBranch) /tmp/authority
	@touch $(flags)/$@

meson: katzenpost-server
	sed -i 's|%%KATZEN_SERVER%%|$(katzenServer):$(katzenBranch)|g' ./plugin/Dockerfile
	docker build -f ./plugin/Dockerfile -t $(mesonServer):$(BRANCH) ./plugin
	sed -i 's|$(katzenServer):$(katzenBranch)|%%KATZEN_SERVER%%|g' ./plugin/Dockerfile
	@touch $(flags)/$@

up: permits up-nonvoting

permits:
	sudo chmod -R 700 ops/nonvoting_testnet/conf/provider?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/mix?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/auth
	@touch $(flags)/$@

up-nonvoting: pull
	GETH_VERSION=$(gethVersion) \
	KATZEN_SERVER=$(katzenpost)\
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml up -d
	@touch $(flags)/$@

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
