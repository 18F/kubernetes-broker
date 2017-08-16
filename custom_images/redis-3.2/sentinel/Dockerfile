FROM debian:jessie-slim

ENV namespace=default
ENV SERVICE=redis-server
ENV SERVICE_PORT=redis-server
ENV REDIS_PASSWORD=redis

ADD ["https://github.com/18f/dig-a/releases/download/v1.0.1/dig-a_linux_amd64", "/dig-a"]
ADD ["https://github.com/18f/dig-srv/releases/download/v1.0.0/dig-srv_linux_amd64", "/dig-srv"]

COPY ["sentinel.template.conf", "k8s-redis-ha-sentinel", "run.sh", "/"]

CMD ["/run.sh"]

ENTRYPOINT ["bash", "-c"]
