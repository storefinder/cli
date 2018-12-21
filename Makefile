TAG?=latest

build-osx:
	docker build --build-arg OS=darwin --build-arg ARCH=amd64 -t storefinder/cli:$(TAG) .
	@docker create --name storefinder  storefinder/cli:$(TAG) \
	&& docker cp storefinder:/usr/bin/storefinder . \
	&& docker rm -f storefinder
	
build-windows:
	docker build --build-arg OS=windows --build-arg ARCH=amd64 -t storefinder/cli:$(TAG) .
	@docker create --name storefinder  storefinder/cli:$(TAG) \
	&& docker cp storefinder:/usr/bin/storefinder . \
	&& docker rm -f storefinder
