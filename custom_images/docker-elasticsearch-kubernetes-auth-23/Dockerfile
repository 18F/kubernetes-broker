FROM quay.io/pires/docker-elasticsearch-kubernetes:2.3.5_1

MAINTAINER devops@gsa.gov

# Install elasticfence plugin
RUN /elasticsearch/bin/plugin install https://raw.githubusercontent.com/elasticfence/elasticsearch-http-user-auth/2.3.5/jar/elasticfence-2.3.5-SNAPSHOT.zip

# Install AWS Cloud plugin
RUN /elasticsearch/bin/plugin install https://download.elastic.co/elasticsearch/release/org/elasticsearch/plugin/cloud-aws/2.3.5/cloud-aws-2.3.5.zip

# Add elasticfence settings to config file
ADD elasticfence.yml /elasticsearch/config/elasticfence.yml
RUN cat /elasticsearch/config/elasticfence.yml >> /elasticsearch/config/elasticsearch.yml
