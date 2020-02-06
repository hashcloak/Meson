TRAVIS_BRANCH ?= $(shell git branch| grep \* | cut -d' ' -f2)
BRANCH=$(TRAVIS_BRANCH)

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

mesonPluginRepo=https://github.com/hashcloak/Meson-plugin
MESON_PLUGIN_BRANCH ?= master
mesonServer=$(dockerRepo)/meson:$(MESON_PLUGIN_BRANCH)

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
	docker pull $(mesonServer) && $(messagePull)$(mesonServer) \
		|| ($(imageNotFound)$(mesonServer) && $(MAKE) build-meson)
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

build-meson:
	git clone $(mesonPluginRepo) /tmp/Meson-plugin || true
	cd /tmp/Meson-plugin && git pull && git checkout $(MESON_PLUGIN_BRANCH)
	make -C /tmp/Meson-plugin build
	@touch $(flags)/$@

genconfig:
	go get github.com/hashcloak/genconfig
	@touch $(flags)/$@

containers: pull genconfig
	KATZEN_AUTH=$(katzenAuth) \
	MESON_IMAGE=$(mesonServer) \
	bash ./ops/containers.sh
	sleep 90
	@touch $(flags)/$@

tests: containers
	bash ./ops/tests.sh

stop:
	bash ./ops/stop.sh
	rm -f $(flags)/containers

logs-auth:
	sudo tail -f /tmp/meson-current/nonvoting/authority.log

logs-providers:
	sudo tail -f /tmp/meson-current/provider-{0,1}/katzenpost.log
