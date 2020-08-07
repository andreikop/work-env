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

RUN useradd ak --shell /bin/zsh --password 'ak'
RUN usermod -G sudo ak
RUN mkdir /home/ak
RUN chown ak:ak /home/ak


RUN apt-get install -y build-essential clang libdbus-1-dev libgtk-3-dev \
                       libnotify-dev libgnome-keyring-dev \
                       libasound2-dev libcap-dev libcups2-dev libxtst-dev \
                       libxss1 libnss3-dev gcc-multilib g++-multilib curl \
                       gperf bison python-dbusmock openjdk-8-jre

ENV LC_ALL=en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8

WORKDIR /home/ak
