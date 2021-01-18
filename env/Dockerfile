FROM ubuntu:20.04

RUN echo 'work-env' > /etc/hostname

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    locales \
    zsh \
    sudo \
    openssl \
    wget \
    curl \
    vim \
    git

RUN locale-gen en_US.UTF-8

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8

COPY work-env-entrypoint.sh /usr/sbin/work-env-entrypoint.sh
ENTRYPOINT /usr/sbin/work-env-entrypoint.sh
