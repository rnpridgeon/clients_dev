ARG BASE_VERSION
FROM ryanp/librd_builder:${BASE_VERSION} as builder

RUN go get github.com/confluentinc/confluent-kafka-go/kafka
WORKDIR /root/go/src/github.com/confluentinc/confluent-kafka-go
COPY ./src/confluent-kafka-go-example/example.go .
RUN go build -o /example example.go 

FROM ryanp/librdkafka:${BASE_VERSION} 

COPY --from=builder /example .

ENTRYPOINT ["./example"]
CMD $@
