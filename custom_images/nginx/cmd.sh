#!/bin/bash

htpasswd -cb /etc/nginx/passwords root "${AUTH_PASSWORD}"

nginx -g 'daemon off;'
