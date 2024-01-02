FROM golang:1.21-bullseye AS build

WORKDIR /qotd

COPY go.mod go.sum ./

RUN --mount=type=cache,target="/root/.cache/go-build" \
  go mod download

COPY app ./app
COPY cmd ./cmd

ENV CGO_ENABLED=0

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" \
  go build -o webserver ./cmd/webserver

RUN chmod +x ./webserver

FROM node:22-alpine AS build-frontend

WORKDIR /qotd

COPY packag*.json ./

RUN npm install

COPY postcss.config.cjs tsconf*.json vite.config.ts index.html .env.* ./
COPY ui ./ui

RUN npm run build

FROM scratch

WORKDIR /

ENV IS_PROD=true

COPY --from=build /qotd/webserver /webserver
COPY --from=build-frontend /qotd/dist /dist

ENTRYPOINT [ "/webserver" ]

EXPOSE 9075