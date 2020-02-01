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

mesonServer=$(dockerRepo)/meson:master


clean:
	rm -rf /tmp/authority
	rm -rf $(flags)

pull: pull-katzen-auth pull-meson
	@touch $(flags)/$@

pull-katzen-auth:
	docker pull $(katzenAuth) && $(messagePull)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && $(MAKE) build-katzen-nonvoting-authority)
	@touch $(flags)/$@

pull-meson:
	docker pull $(mesonServer)
	@touch $(flags)/$@

push: push-katzen-auth

push-katzen-auth:
	docker push $(katzenAuth) && $(messagePush)$(katzenAuth) \
		|| ($(imageNotFound)$(katzenAuth) && \
				$(MAKE) build-katzen-nonvoting-authority && \
				docker push $(katzenAuth))

build: build-katzen-nonvoting-authority

build-katzen-nonvoting-authority:
	 git clone $(katzenAuthRepo) /tmp/auth || git --git-dir=/tmp/auth/.git --work-tree=/tmp/auth pull origin master
	docker build -f /tmp/auth/Dockerfile.nonvoting -t $(katzenAuth) /tmp/auth
	@touch $(flags)/$@

genconfig:
	go get github.com/hashcloak/genconfig
	genconfig -o ops/conf -n 3
	@touch $(flags)/$@

integration-tests: pull genconfig
	MESON_IMAGE=$(mesonServer) \
	KATZEN_AUTH=$(katzenAuth) \
	docker-compose -f ./ops/integration-test.compose.yml up -d

down-integration:
	docker-compose -f ./ops/integration-test.compose.yml down
