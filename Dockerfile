FROM golang:1.16-buster AS builder

WORKDIR /scmdhttpd/
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN go build

FROM debian:buster-slim
LABEL maintainer="Andreas Schulze"

COPY --from=builder /scmdhttpd/scmdhttpd /
COPY entrypoint /
RUN chmod 0555 /entrypoint

ENTRYPOINT ["/entrypoint"]
