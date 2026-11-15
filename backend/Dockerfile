# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/tex-api ./src/cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
  && addgroup -S app \
  && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/tex-api /app/tex-api
COPY VERSION /app/VERSION

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/tex-api"]
