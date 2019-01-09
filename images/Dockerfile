FROM alpine:3.8
ARG VERSION

RUN apk add --update --no-cache git bash build-base \
	openssl-dev cyrus-sasl-dev lz4-dev \
	python2-dev py-setuptools \
	go 

RUN git clone https://github.com/edenhill/librdkafka.git
WORKDIR /librdkafka
RUN git checkout ${VERSION} && ./configure && make && make install
WORKDIR /
