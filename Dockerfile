FROM node:24-alpine AS frontend

COPY . /app
WORKDIR /app
RUN corepack enable
RUN pnpm install --filter shopmon-frontend
RUN cd frontend && pnpm run build

FROM node:24-alpine AS api

COPY . /app
WORKDIR /app
RUN corepack enable
RUN pnpm install --filter shopmon
RUN rm -rf /app/frontend

FROM node:24-alpine AS final

COPY --from=api /app/ /app/
COPY --from=frontend /app/frontend/dist /app/api/dist/

WORKDIR /app/api
ENV NODE_ENV=production
EXPOSE 3000
ENTRYPOINT [ "node", "--no-warnings=ExperimentalWarning" ]

CMD [ "src/index.ts" ]
