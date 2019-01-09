ARG BASE_VERSION
FROM ryanp/librd_builder:${BASE_VERSION} as builder
ARG VERSION

RUN git clone https://github.com/confluentinc/confluent-kafka-python.git
WORKDIR /confluent-kafka-python
RUN git checkout ${VERSION} && python ./setup.py build && python ./setup.py install
WORKDIR /

FROM ryanp/librdkafka:${BASE_VERSION}

COPY --from=builder /usr/lib/python2.7 /usr/lib/python2.7
COPY --from=builder /usr/lib/libpython2.7.so.1.0 /usr/lib/libpython2.7.so.1.0
COPY --from=builder /usr/bin/python2 /usr/bin/python

