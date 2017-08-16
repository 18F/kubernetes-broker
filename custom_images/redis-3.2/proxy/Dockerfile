FROM debian:jessie-slim

ENV LISTEN_ADDRESS=:6379
ENV SENTINEL_ADDRESS=sentinel:26379
ENV REDIS_MASTER_NAME=master

ADD ["https://github.com/18f/redis-sentinel-proxy/releases/download/v1.0.1/redis-sentinel-proxy_linux_amd64", "/redis-sentinel-proxy"]

COPY ["k8s-redis-sentinel-proxy", "run.sh", "/"]

CMD ["/run.sh"]

ENTRYPOINT ["bash", "-c"]
