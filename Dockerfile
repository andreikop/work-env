FROM ubuntu:20.04

RUN echo 'docker-env' > /etc/hostname

RUN apt-get update && apt-get install -y \
	locales \
	zsh \
	wget \
	sudo \
	git \
	vim

RUN locale-gen en_US.UTF-8

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8

RUN useradd ak --shell /bin/zsh --password 'ak'
RUN usermod -G sudo ak
RUN mkdir /home/ak
RUN chown ak:ak /home/ak


RUN DEBIAN_FRONTEND=noninteractive apt-get install -y build-essential meson

RUN DEBIAN_FRONTEND=noninteractive apt-get install -y \
  gstreamer1.0-tools \
  libgstreamer1.0-dev \
  libgstreamer-plugins-bad1.0-dev \
  gstreamer1.0-rtsp \
  libgstrtspserver-1.0-dev \
  cmake


WORKDIR /home/ak
