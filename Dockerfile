FROM ubuntu:20.04

RUN echo 'work-env' > /etc/hostname

RUN apt-get update && apt-get install -y \
	locales \
	zsh \
	sudo

RUN locale-gen en_US.UTF-8

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8

RUN useradd \
    --create-home \
    --shell /bin/zsh \
    --groups sudo \
    --uid 1000  \
    --password "$(openssl passwd -1 ak)" \
    ak


RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y wget git vim

# RUN DEBIAN_FRONTEND=noninteractive \
#    apt-get install -y build-essential meson

# RUN DEBIAN_FRONTEND=noninteractive apt-get install -y \
#   gstreamer1.0-tools \
#   libgstreamer1.0-dev \
#   libgstreamer-plugins-bad1.0-dev \
#   gstreamer1.0-rtsp \
#   libgstrtspserver-1.0-dev \
#   cmake

USER ak

WORKDIR /home/ak
