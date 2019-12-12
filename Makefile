
flags=.makeFlags
VPATH=$(flags)
$(shell mkdir -p $(flags))
gethVersion=v1.9.8
.PHONY:up down

all: hashcloak-geth katzenpost-server meson katzenpost-nonvoting-authority

pull:
	git clone https://github.com/katzenpost/server /tmp/server || true
	git clone https://github.com/katzenpost/authority /tmp/authority || true
	git clone https://github.com/hashcloak/Meson /tmp/Meson || true
	@touch $(flags)/$@

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

hashcloak-geth:
	sed -i 's|%%GETH_VERSION%%|$(gethVersion)|g' ./ops/geth.Dockerfile
	docker build -f ./ops/geth.Dockerfile -t hashcloak/client-go:$(gethVersion) .
	sed -i 's|$(gethVersion)|%%GETH_VERSION%%|g' ./ops/geth.Dockerfile
	@touch $(flags)/$@

katzenpost-server: pull
	docker build -f /tmp/server/Dockerfile -t katzenpost/server /tmp/server
	@touch $(flags)/$@

katzenpost-voting-authority: pull
	docker build -f /tmp/authority/Dockerfile.voting -t katzenpost/voting_authority /tmp/authority
	@touch $(flags)/$@

katzenpost-nonvoting-authority: pull
	docker build -f /tmp/authority/Dockerfile.nonvoting -t katzenpost/nonvoting_authority /tmp/authority
	@touch $(flags)/$@

meson:
	docker build -f ./plugin/Dockerfile -t hashcloak/meson ./plugin
	@touch $(flags)/$@

up: permits up-nonvoting

permits:
	sudo chmod -R 700 ops/nonvoting_testnet/conf/provider?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/mix?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/auth
	sudo chmod -R 700 ops/nonvoting_testnet/conf/goerli
	sudo chmod -R 700 ops/nonvoting_testnet/conf/rinkeby
	@touch $(flags)/$@

up-nonvoting: permits all
	GETH_VERSION=$(gethVersion) \
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml up -d

down: down-nonvoting

down-nonvoting: 
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml down

rebuild: rebuild-meson

rebuild-meson:
	docker build -f ./Dockerfile -t hashcloak/meson .
