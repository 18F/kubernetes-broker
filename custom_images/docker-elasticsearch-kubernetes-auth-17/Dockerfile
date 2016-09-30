FROM quay.io/pires/docker-elasticsearch-kubernetes:1.7.2

MAINTAINER devops@gsa.gov

# Set environment
ENV CLUSTER_NAME ${CLUSTER_NAME:-elasticsearch-default}
ENV NODE_MASTER ${NODE_MASTER:-true}
ENV NODE_DATA ${NODE_DATA:-true}
ENV HTTP_ENABLE ${HTTP_ENABLE:-true}
ENV MULTICAST ${MULTICAST:-false}

ENV NAMESPACE ${NAMESPACE:-default}
ENV DISCOVERY_SERVICE ${DISCOVERY_SERVICE:-elasticsearch-discovery}

# Install basic auth plugin
ENV HB_VERSION=1.5.1
ENV HTTP_BASIC_URL https://github.com/Asquera/elasticsearch-http-basic/releases/download/v${HB_VERSION}/elasticsearch-http-basic-${HB_VERSION}.jar
RUN /elasticsearch/bin/plugin \
  --url $HTTP_BASIC_URL \
  --install http-basic \
  --silent

# Add elasticfence settings to config file
ADD basicauth.yml /elasticsearch/config/basicauth.yml
RUN cat /elasticsearch/config/basicauth.yml >> /elasticsearch/config/elasticsearch.yml
