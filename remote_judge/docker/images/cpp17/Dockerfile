FROM alpine:3.19

RUN apk add --no-cache g++ make libstdc++ \
    && adduser -D -u 1000 runner

WORKDIR /workspace

USER runner
