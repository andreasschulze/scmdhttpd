# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 2.3.3

- add security header also when responding with a redirect
- updated the way to disable http2
- updated required go- and module versions

## 2.3.2

- entrypoint support pass-through ENV[GODEBUG]
  documentation say
  "... [is a comma-separated list of name=val pairs](https://pkg.go.dev/runtime#hdr-Environment_Variables) ..."

## 2.3.1

- Referrer-policy changed to 'no-referrer' as suggested by internet.nl

## 2.3.0

- RFC 9116 support
- FIX: fetching ACME certs was broken, don't use 2.2.x!

## 2.2.1

- updated Github workflows
- updated dependenies

## 2.2.0

- updated Github workflows
- use go-1.21.x
- fix some dependenies
- require golang.org/x/crypto v0.17.0

## 2.1.0

- valid Cache-Control header

## 2.0.7

- log TLS version and cipher

## 2.0.6

- log 404 errors, too

## 2.0.5

- use go-1.18
- minor updates on buildfiles and documentation

## 2.0.4

- Apache-2.0 License (as the code heavily based on github.com/danmarg/sts-mate)
- Github Workflow: shellcheck
- Github Workflow: markdownlint
- SECURITY: escape new line characters in referer and user-agent before writing
  to stdout/log

## 2.0.3

- `--version` implemented
- module dependencies update
- src rewritten as suggested by gofmt
