FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/minicord-backend ./cmd/server

FROM alpine:3.22
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/minicord-backend /usr/local/bin/minicord-backend
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/minicord-backend"]
