# notifier

Asynchronous HTTP notification client and CLI written in Go.

This project implements a small library that sends HTTP POST notifications
to a configured endpoint. The API is designed to be non-blocking and to
handle spikes in incoming messages without exhausting system resources.

A small CLI (`notify`) will be provided to demonstrate the library. It
reads messages from stdin and sends them to a configured URL at a
configurable interval.

## Status

Work in progress.

## Development

Run tests:

```
go test ./...
```
Build the CLI:
```
go build ./cmd/notify
```
