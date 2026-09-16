FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/minicord-backend ./cmd/server

FROM alpine:3.22
RUN adduser -D -H appuser && mkdir -p /data && chown appuser:appuser /data
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=build /out/minicord-backend /usr/local/bin/minicord-backend
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
