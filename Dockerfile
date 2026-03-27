# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS frontend-build

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build && cp -R dist/. /out/

FROM golang:1.26-alpine AS api-build

WORKDIR /src/api
COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./
COPY --from=frontend-build /out /tmp/frontend-dist
RUN mkdir -p /src/api/internal/webui/dist && \
	cp -R /tmp/frontend-dist/. /src/api/internal/webui/dist/ && \
	CGO_ENABLED=0 go build -ldflags="-s -w" -o /shopmon .

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=api-build /shopmon /usr/local/bin/shopmon

EXPOSE 8080
ENTRYPOINT ["shopmon"]
CMD ["server"]
