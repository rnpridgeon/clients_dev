FROM ryanp/librd_builder:0.11.5 as builder

ADD ./src/librdkafka-example/example.c /

RUN gcc -I/usr/local/include/librdkafka example.c -o rdkafka_example /usr/local/lib/librdkafka.so \
	-lrdkafka -lz -lpthread -lrt

FROM ryanp/librdkafka:0.11.5
COPY --from=builder /rdkafka_example .

ENTRYPOINT ["./rdkafka_example"]
