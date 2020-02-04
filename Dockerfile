FROM ubuntu:18.04

RUN echo 'docker-env' > /etc/hostname

RUN apt-get update && apt-get install -y zsh wget sudo git

RUN useradd ak --shell /bin/zsh --password ''
RUN usermod -G sudo ak
RUN mkdir /home/ak
RUN chown ak:ak /home/ak
WORKDIR /home/ak
