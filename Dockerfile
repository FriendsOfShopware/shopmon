# syntax=docker/dockerfile:1.7

ARG BUILDPLATFORM=linux/amd64
ARG TARGETPLATFORM=linux/amd64
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT

FROM --platform=$BUILDPLATFORM node:26-alpine AS frontend-build

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build && cp -R dist/. /out/

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS api-build

WORKDIR /src/api
COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./
COPY --from=frontend-build /out /tmp/frontend-dist
RUN mkdir -p /src/api/internal/webui/dist && \
	cp -R /tmp/frontend-dist/. /src/api/internal/webui/dist/ && \
	GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /shopmon .

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian13:nonroot

COPY --from=api-build /shopmon /usr/local/bin/shopmon

EXPOSE 8080
ENTRYPOINT ["shopmon"]
CMD ["server"]
