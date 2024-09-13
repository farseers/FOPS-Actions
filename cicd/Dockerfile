FROM alpine:latest AS base
RUN echo "export PATH=$PATH:/usr/local/go/bin" >> /root/.bashrc
RUN echo "export PS1='\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$'" >> /root/.bashrc
RUN echo "export GOPROXY=https://goproxy.cn,direct" >> /root/.bashrc
RUN echo "alias ll='ls -l'" >> /root/.bashrc
RUN source /root/.bashrc
RUN apk add git
RUN apk add bash
RUN apk add docker
RUN apk add curl