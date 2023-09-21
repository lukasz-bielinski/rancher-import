.PHONY: build push build-push build-push-no-cache

IMAGE_NAME := lukaszbielinski/rancher-import

build:
		docker build . -t $(IMAGE_NAME)

build-no-cache:
		docker build --no-cache . -t $(IMAGE_NAME)

push: build
		docker push $(IMAGE_NAME)

push-no-cache: build-no-cache
		docker push $(IMAGE_NAME)

build-push: build push

build-push-no-cache: build-no-cache push
