ARG BASE_VERSION
FROM ryanp/librd_builder:$BASE_VERSION as builder

FROM alpine:3.8 
RUN apk --no-cache add ca-certificates 

COPY --from=builder /lib/libssl.so.1.0.0 /usr/lib
COPY --from=builder /lib/libcrypto.so.1.0.0 /usr/lib

COPY --from=builder /usr/lib/libsasl2.so.3.0.0 /usr/lib/libsasl2.so.3

COPY --from=builder /usr/lib/liblz4.so.1.8.2 /usr/lib/liblz4.so.1

COPY --from=builder /usr/local/lib/librdkafka.so.1 /lib
COPY --from=builder /usr/local/include/librdkafka/rdkafka.h /usr/include/rdkafka.h
