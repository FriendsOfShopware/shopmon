FROM oven/bun:latest AS frontend

COPY . /app
WORKDIR /app
RUN bun install
RUN cd frontend && bun run build

FROM oven/bun:latest AS api

COPY . /app
WORKDIR /app
RUN bun install
RUN cd api && bun run build:app
RUN cd api && bun run build:cron

FROM oven/bun:latest AS final

COPY --from=frontend /app/frontend/dist /app/dist/
COPY --from=api /app/api/src/app*.js* /app/
COPY --from=api /app/api/src/cron/cron*.js* /app/

WORKDIR /app
RUN bun install sharp
ENV NODE_ENV=production
EXPOSE 3000
ENTRYPOINT [ "bun" ]

STOPSIGNAL SIGKILL

CMD [ "app.js" ]
