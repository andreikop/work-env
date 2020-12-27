FROM ubuntu:20.04

RUN echo 'work-env' > /etc/hostname

RUN apt-get update && apt-get install -y \
    locales \
    zsh \
    sudo \
    openssl

RUN locale-gen en_US.UTF-8

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8


ARG WORK_ENV_USER_SHELL=/bin/bash
ARG WORK_ENV_USER_ID=1000
ARG WORK_ENV_USER_NAME=ak
ARG WORK_ENV_USER_PASSWORD=ak

RUN useradd \
    --create-home \
    --shell ${WORK_ENV_USER_SHELL} \
    --groups sudo \
    --uid ${WORK_ENV_USER_ID}  \
    --password "$(openssl passwd -1 ${WORK_ENV_USER_PASSWORD})" \
    ${WORK_ENV_USER_NAME}


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

USER ${WORK_ENV_USER_NAME}

WORKDIR /home/ak
