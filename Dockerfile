FROM ubuntu:18.04

RUN echo 'docker-env' > /etc/hostname

RUN apt-get update && apt-get install -y \
	locales \
	zsh \
	wget \
	sudo \
	git \
	vim

RUN locale-gen en_US.UTF-8

RUN useradd ak --shell /bin/zsh --password ''
RUN usermod -G sudo ak
RUN mkdir /home/ak
RUN chown ak:ak /home/ak

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8

WORKDIR /home/ak