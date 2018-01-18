FROM nginx:1.13.8

RUN apt-get update \
  && apt-get install -y apache2-utils curl \
  && rm -rf /var/lib/apt/lists/*

ADD cmd.sh /cmd.sh
ADD nginx.conf /etc/nginx/nginx.conf

CMD ["/cmd.sh"]
