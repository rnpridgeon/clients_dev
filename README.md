# clients_dev

This repository makes no attempt to follow docker best practices nor does it intend to. I am more than open to patches however if someone more familiar with docker wishes to contribute. 

## usage notes:

If a `VERSION` is not specified images are built from the master branch. 
If no recipe is provided(`client` in the usage section) confluent-kafka-python and confluent-kafka-go will be built. 
Librdkafka is always built since it underpins the other two clients

## usage: 

make [VERSION] [client]

## example:

make VERSION=v0.11.6 confluent-kafka-python

