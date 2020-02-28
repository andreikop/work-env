IMAGE := docker-env
CONTAINER_NAME := env

all:
	docker build --tag $(IMAGE) .

run:
	-docker run \
		--interactive --tty \
		--name $(CONTAINER_NAME) \
		--volume /home:/home \
		--volume /tmp:/tmp \
		--user `id -u` \
		$(IMAGE) \
		/bin/zsh

attach:
	-docker start -ai $(CONTAINER_NAME)

rm-container:
	docker rm $(CONTAINER_NAME)
