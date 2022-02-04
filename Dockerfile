FROM golang:1.17-bullseye AS builder

WORKDIR /scmdhttpd/
COPY go.mod go.sum *.go ./
RUN go mod download \
 && go build \
 && strip scmdhttpd

FROM debian:bullseye-slim
COPY --from=builder /scmdhttpd/scmdhttpd /
COPY entrypoint /
RUN chmod 0555 /entrypoint

ENTRYPOINT ["/entrypoint"]
