DIR=$(shell pwd)
VERSION_TAG?=$(subst v,,$(VERSION))

ifeq ($(origin VERSION), undefined)
  VERSION=master
  VERSION_TAG=latest
endif

all: confluent-kafka-python-example confluent-kafka-go-example

librd_builder: $(DIR)/images/Dockerfile
	docker build --build-arg VERSION=$(VERSION) --build-arg=BASE_VERSION=$(VERSION_TAG) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/Dockerfile . 

librdkafka: librd_builder $(DIR)/images/librdkafka/Dockerfile
	docker build --build-arg VERSION=$(VERSION) --build-arg=BASE_VERSION=$(VERSION_TAG) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/$@/Dockerfile .

confluent-kafka-python: librdkafka $(DIR)/images/confluent-kafka-python/Dockerfile
	docker build --build-arg VERSION=$(VERSION) --build-arg=BASE_VERSION=$(VERSION_TAG) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/$@/Dockerfile .

ifeq ('$(VERSION)', 'v0.11.5')
librdkafka-example: librdkafka $(DIR)/images/librdkafka-example/Dockerfile \
  $(DIR)/src/librdkafka-example/example.c $(DIR)/src/conf/producer.properties \
  $(DIR)/src/conf/consumer.properties
	docker build --build-arg VERSION=$(VERSION) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/$@/Dockerfile . 
endif 

confluent-kafka-python-example: librdkafka confluent-kafka-python \
  $(DIR)/images/confluent-kafka-python-example/Dockerfile \
  $(DIR)/src/conf/producer.properties $(DIR)/src/conf/consumer.properties
	docker build --build-arg VERSION=$(VERSION) --build-arg=BASE_VERSION=$(VERSION_TAG) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/$@/Dockerfile .	

confluent-kafka-go-example: librdkafka \
  $(DIR)/images/confluent-kafka-go-example/Dockerfile \
  $(DIR)/src/conf/producer.properties $(DIR)/src/conf/consumer.properties
	docker build --build-arg=BASE_VERSION=$(VERSION_TAG) \
		-t ryanp/$@:$(VERSION_TAG) -f $(DIR)/images/$@/Dockerfile .	

clean:
	docker-compose down -v --remove-orphans --rmi all
#	docker system prune -a
#	docker image prune
#	docker volume prune
#	docker network prune
