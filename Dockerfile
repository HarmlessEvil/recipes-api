FROM golang:1.20-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /recipes-api

FROM gcr.io/distroless/base-debian11

ARG API_VERSION
ENV API_VERSION=$API_VERSION

WORKDIR /

COPY --from=build-stage /recipes-api /recipes-api

USER nonroot:nonroot

ENTRYPOINT ["/recipes-api"]
