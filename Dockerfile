FROM golang:1.21-bookworm AS builder

WORKDIR /scmdhttpd/
COPY go.mod go.sum *.go ./
RUN go mod download \
 && go build \
 && strip scmdhttpd

FROM debian:bullseye-slim
COPY --from=builder /scmdhttpd/scmdhttpd /
COPY entrypoint /
RUN apt-get update \
 && apt-get install --no-install-recommends --no-install-suggests -y ca-certificates=20210119 \
 && rm -rf /var/lib/apt/lists/* \
 && chmod 0555 /entrypoint

ENTRYPOINT ["/entrypoint"]
