FROM alpine:latest AS base
RUN echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
RUN echo "export PS1='\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$'" >> /etc/profile
RUN echo "alias ll='ls -l'" >> /etc/profile
RUN source /etc/profile
RUN apk add git
RUN apk add bash
RUN apk add docker
