FROM golang:1.25-alpine AS sqlmigrate
RUN apk add --no-cache gcc musl-dev sqlite-dev
ENV CGO_ENABLED=1
RUN go install github.com/rubenv/sql-migrate/sql-migrate@v1.8.1

FROM golang:1.25-alpine AS build
ARG VERSION=dev
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /src
COPY service/go.mod service/go.sum ./
RUN go mod download
COPY service/ ./
RUN CGO_ENABLED=1 go build \
  -ldflags "-s -w -X github.com/jamesread/starapp/service/internal/buildinfo.Version=${VERSION}" \
  -o /starapp .

FROM alpine:3.20
LABEL org.opencontainers.image.source="https://github.com/jamesread/StarLoom"
LABEL org.opencontainers.image.title="StarApp"
LABEL org.opencontainers.image.description="Family star rewards — parents award stars; children redeem privileges."
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=sqlmigrate /go/bin/sql-migrate /usr/bin/sql-migrate
COPY --from=build /starapp /usr/bin/starapp
COPY database /var/app/database
COPY frontend/dist /usr/share/starapp/webui
COPY config/container/config.yaml /config/config.yaml
COPY docker-entrypoint.sh /usr/local/bin/starapp-entrypoint.sh
RUN chmod +x /usr/local/bin/starapp-entrypoint.sh /usr/bin/starapp /usr/bin/sql-migrate
ENV DB_DRIVER=sqlite
EXPOSE 8080
VOLUME /config
ENTRYPOINT ["/usr/local/bin/starapp-entrypoint.sh"]
