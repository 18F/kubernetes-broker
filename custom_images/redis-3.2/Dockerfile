FROM redis:3.2.10

RUN mkdir -p /opt/bin

ADD https://github.com/18f/dig-a/releases/download/v1.0.1/dig-a_linux_amd64 /opt/bin/dig-a
ADD https://github.com/18f/dig-srv/releases/download/v1.0.0/dig-srv_linux_amd64 /opt/bin/dig-srv
ADD https://github.com/18F/redis-sentinel-proxy/releases/download/v1.0.2/redis-sentinel-proxy_linux_amd64 /opt/bin/redis-sentinel-proxy

COPY redis.template.conf sentinel.template.conf /opt/

COPY bootstrap.sh k8s-redis-sentinel-proxy k8s-redis-ha-sentinel k8s-redis-ha-server k8s-redis-ha-sentinel-ready k8s-redis-ha-server-ready pre-stop.sh /opt/bin/

RUN chmod -R +x /opt/bin
