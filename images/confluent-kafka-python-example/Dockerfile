ARG BASE_VERSION
FROM ryanp/confluent-kafka-python:${BASE_VERSION}

ADD ./src/confluent-kafka-python-example/example.py . 
ADD ./src/confluent-kafka-python-example/tests/testconf.json /tests/testconf.json

ENTRYPOINT ["python", "example.py" ]
