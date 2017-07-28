FROM 18fgsa/elasticsearch:2.4.4

RUN bin/plugin install io.fabric8/elasticsearch-cloud-kubernetes/2.4.4
RUN bin/plugin install https://raw.githubusercontent.com/elasticfence/elasticsearch-http-user-auth/2.4.4/jar/elasticfence-2.4.4-SNAPSHOT.zip
RUN bin/plugin install https://download.elastic.co/elasticsearch/release/org/elasticsearch/plugin/cloud-aws/2.4.4/cloud-aws-2.4.4.zip

ENV BOOTSTRAP_MLOCKALL=false NODE_DATA=true NODE_MASTER=true ES_JAVA_OPTS="-Djava.net.preferIPv4Stack=true -Dmapper.allow_dots_in_name=true"


# pre-stop-hook.sh and dependencies
RUN apt-get update && apt-get install -y \
    jq \
    curl \
  && rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /entrypoint.sh

COPY pre-stop-hook.sh /pre-stop-hook.sh

COPY elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml

ENTRYPOINT ["/entrypoint.sh"]
CMD ["elasticsearch"]
