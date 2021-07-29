FROM golang:1.16.2-stretch as drago-builder

RUN mkdir /data
WORKDIR /data
VOLUME /data

RUN curl -sS http://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
    echo "deb http://dl.yarnpkg.com/debian/ stable main" |  tee /etc/apt/sources.list.d/yarn.list && \
    curl -sL http://deb.nodesource.com/setup_15.x | bash - && \
    apt-get install -y nodejs && \
    apt-get update && \
    apt-get remove cmdtest && \
    apt-get install -y yarn

RUN apt-get install -y gcc-arm-linux-gnueabihf libc6-dev-armhf-cross \
                       gcc-aarch64-linux-gnu libc6-dev-arm64-cross
