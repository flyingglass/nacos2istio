FROM alpine:3.13.5

ENV NACOS_ADDR=""
ENV NACOS_WATCH_NS="default"
ENV MODE="se"

COPY nacos2istio /usr/local/bin/
ENTRYPOINT /usr/local/bin/nacos2istio -addr=$NACOS_ADDR -ns=$NACOS_WATCH_NS -mode=$MODE