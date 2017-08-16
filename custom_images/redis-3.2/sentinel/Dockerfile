FROM debian:jessie-slim

ADD ["https://github.com/18f/dig-a/releases/download/v1.0.1/dig-a_linux_amd64", "/dig-a"]
ADD ["https://github.com/18f/dig-srv/releases/download/v1.0.0/dig-srv_linux_amd64", "/dig-srv"]

COPY ["sentinel.template.conf", "k8s-redis-ha-sentinel", "run.sh", "/"]

RUN apt-get update && apt-get install -y \
  netcat

CMD ["/run.sh"]

ENTRYPOINT ["bash", "-c"]
