FROM debian:jessie-slim

ENV LISTEN_ADDRESS=:6379
ENV SENTINEL_ADDRESS=sentinel:26379
ENV REDIS_MASTER_NAME=master

RUN mkdir -p /opt/bin

ADD ["https://github.com/18f/redis-sentinel-proxy/releases/download/v1.0.1/redis-sentinel-proxy_linux_amd64", "/opt/bin/redis-sentinel-proxy"]

COPY ["k8s-redis-sentinel-proxy", "/opt/bin"]

RUN chmod -R +x /opt/bin

CMD ["/opt/bin/k8s-redis-sentinel-proxy"]

ENTRYPOINT ["bash", "-c"]
