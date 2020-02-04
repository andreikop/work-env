IMAGE := docker-env
CONTAINER_NAME := env

all:
	docker build --tag $(IMAGE) .

run:
	-docker run \
		--interactive --tty \
		--name $(CONTAINER_NAME) \
		--volume $(HOME):/home/ak \
		--user `id -u` \
		$(IMAGE) \
		/bin/zsh

rm-container:
	docker rm $(CONTAINER_NAME)
