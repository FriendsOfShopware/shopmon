FROM oven/bun:latest AS frontend

COPY . /app
WORKDIR /app
RUN bun install
RUN cd frontend && bun run build

FROM oven/bun:latest AS api

COPY . /app
WORKDIR /app
RUN bun install
RUN rm -rf /app/frontend

FROM oven/bun:latest AS final

ARG SENTRY_RELEASE="unknown"
ENV SENTRY_RELEASE=${SENTRY_RELEASE}

COPY --from=api /app/ /app/
COPY --from=frontend /app/frontend/dist /app/api/dist/

WORKDIR /app/api
ENV NODE_ENV=production
EXPOSE 3000
ENTRYPOINT [ "bun" ]

STOPSIGNAL SIGKILL

CMD [ "src/index.ts" ]
