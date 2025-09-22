ARG BIN_NAME=mqtt2ntfy
ARG BIN_VERSION=<unknown>

FROM golang:1-alpine AS builder
ARG BIN_NAME
ARG BIN_VERSION

RUN apk --no-cache add ca-certificates

WORKDIR /src/mqtt2ntfy
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME} .

FROM scratch
ARG BIN_NAME
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/mqtt2ntfy/out/${BIN_NAME} /usr/bin/mqtt2ntfy
ENTRYPOINT ["/usr/bin/mqtt2ntfy"]

LABEL license="GPL3"
LABEL maintainer="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.authors="Chris Dzombak <https://www.dzombak.com>"
LABEL org.opencontainers.image.url="https://github.com/cdzombak/mqtt2ntfy"
LABEL org.opencontainers.image.documentation="https://github.com/cdzombak/mqtt2ntfy/blob/main/README.md"
LABEL org.opencontainers.image.source="https://github.com/cdzombak/mqtt2ntfy.git"
LABEL org.opencontainers.image.version="${BIN_VERSION}"
LABEL org.opencontainers.image.licenses="GPL3"
LABEL org.opencontainers.image.title="${BIN_NAME}"
LABEL org.opencontainers.image.description="MQTT to Ntfy bridge"