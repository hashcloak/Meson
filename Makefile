flags=.makeFlags
VPATH=$(flags)
$(shell mkdir -p $(flags))

dockerRepo=hashcloak
katzenAuthRepo=https://github.com/katzenpost/authority
katzenAuthTag=$(shell git ls-remote --heads https://github.com/katzenpost/authority  | grep master | cut -c1-7)
katzenAuth=$(dockerRepo)/katzenpost-auth:$(katzenAuthTag)

messagePush=echo "LOG: Image already exists in docker.io/$(dockerRepo). Not pushing: "
messagePull=echo "LOG: Success in pulling image: "
imageNotFound=echo "LOG: Image not found... building: "

clean:
	rm -rf /tmp/authority
	rm -rf $(flags)

clean-data:
	rm -rf ./ops/nonvoting_testnet/conf/provider?
	rm -rf ./ops/nonvoting_testnet/conf/mix?
	rm -rf ./ops/nonvoting_testnet/conf/auth
	git checkout ./ops/nonvoting_testnet/conf
	rm -r $(flags)/permits || true
	$(MAKE) permits

pull: pull-katzen-auth pull-meson

pull-katzen-auth:
	docker pull $(katzenAuth) && $(messagePull)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && $(MAKE) build-katzen-nonvoting-authority)

pull-meson:
	docker pull $(mesonServer):master

push: push-katzen-auth

push-katzen-auth:
	docker push $(katzenAuth) && $(messagePush)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && \
				$(MAKE) build-katzen-nonvoting-authority && \
				docker push $(katzenAuth))

build: build-katzen-nonvoting-authority

build-katzen-nonvoting-authority:
	git clone $(katzenAuthRepo) /tmp/authority || true
	docker build -f /tmp/authority/Dockerfile.nonvoting -t $(katzenAuth) /tmp/authority
	@touch $(flags)/$@

permits:
	sudo chmod -R 700 ops/nonvoting_testnet/conf/provider?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/mix?
	sudo chmod -R 700 ops/nonvoting_testnet/conf/auth
	@touch $(flags)/$@

up-nonvoting: permits pull up-nonvoting
	KATZEN_SERVER=$(katzenServer) \
	KATZEN_AUTH=$(katzenAuth) \
	MESON_IMAGE=$(mesonServer):$(BRANCH) \
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml up -d
	@touch $(flags)/$@

down:
	docker-compose -f ./ops/nonvoting_testnet/docker-compose.yml down

integration-tests: up-nonvoting

