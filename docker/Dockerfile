FROM node:18-alpine3.17

WORKDIR /app

COPY ./docker/entrypoint.sh entrypoint.sh

RUN chmod +x entrypoint.sh

COPY ./docker/tool.sh tool.sh

RUN chmod +x tool.sh

COPY ./docker/nginx.conf /etc/nginx/http.d/default.conf

COPY ./api /app/api
COPY ./frontend /app/frontend
COPY ./shared /app/shared

COPY ./docker/db.sql /app/db.sql

RUN mkdir -p /app/data/db

# Frontend preparation

WORKDIR /app/frontend

RUN npm install

RUN apk add --no-cache nginx sqlite apache2-utils

RUN npm run build

RUN mv dist /var/www

# Api preparation

WORKDIR /app/api

RUN npm install

RUN npm install -g miniflare

VOLUME /app/data

# cleanup unused things

RUN rm -rf /app/frontend

WORKDIR /app

ENTRYPOINT ["/app/entrypoint.sh"]
