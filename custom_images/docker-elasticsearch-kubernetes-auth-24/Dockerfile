FROM quay.io/pires/docker-elasticsearch-kubernetes:2.4.1

MAINTAINER devops@gsa.gov

# Append to ES_JAVA_OPTS. NOTE: Nothing else is ES_JAVA_OPTS for now but in
# future versions, there will be.
# https://github.com/pires/docker-elasticsearch/blob/master/Dockerfile#L27
# This allows ES 2.4 to have dots in names
# More here: https://www.elastic.co/guide/en/elasticsearch/reference/2.4/dots-in-names.html
ENV ES_JAVA_OPTS="${ES_JAVA_OPTS} -Dmapper.allow_dots_in_name=true"

# Install elasticfence plugin
RUN /elasticsearch/bin/plugin install https://raw.githubusercontent.com/elasticfence/elasticsearch-http-user-auth/2.4.1/jar/elasticfence-2.4.1-SNAPSHOT.zip

# Install AWS Cloud plugin
RUN /elasticsearch/bin/plugin install https://download.elastic.co/elasticsearch/release/org/elasticsearch/plugin/cloud-aws/2.4.1/cloud-aws-2.4.1.zip
ADD logging.aws.yml /elasticsearch/config/logging.aws.yml

# Add elasticfence settings to config file
ADD elasticfence.yml /elasticsearch/config/elasticfence.yml
RUN cat /elasticsearch/config/elasticfence.yml >> /elasticsearch/config/elasticsearch.yml
