# notifier

Asynchronous HTTP notification client and CLI written in Go.

This project implements a small library that sends HTTP POST notifications
to a configured endpoint. The API is designed to be non-blocking and to
handle spikes in incoming messages without exhausting system resources.

A small CLI (`notify`) is provided to demonstrate the library. It
reads messages from `stdin` and sends them to a configured URL at a
configurable interval.

A basic HTTP server is provided to exercise the library and CLI.

---

## Features

- Non-blocking notification API
- Bounded queue to provide backpressure
- Worker pool to control concurrency
- HTTP client with connection reuse
- Graceful shutdown support

---

## CLI

A simple executable is provided under `cmd/notify`.

### Build

```bash
go build ./cmd/notify
```

### Example

```bash
echo "hello world" | ./notify --url http://localhost:8080/notify
```

You can display the available options with:

```bash
notify --help
```

### Flags

| Flag         | Description                 | Default  |
| ------------ | --------------------------- | -------- |
| `--url`      | Notification endpoint       | required |
| `--interval` | Delay between notifications | `5s`     |
| `--workers`  | Worker pool size            | `8`      |
| `--queue`    | Queue capacity              | `1024`   |
| `--timeout`  | Request timeout             | `10s`    |

Example sending multiple lines:

```bash
cat messages.txt | ./notify --url http://localhost:8080/notify --interval 1s
```

---

## Tests

Run all tests with:

```bash
go test ./...
```
---

## Local Test Server

For manual end-to-end testing a basic HTTP server is provided under `cmd/testserver`.

It accepts `POST /notify`, logs the request body, and returns a configurable 
status code. The test server can delay responses to simulate slower services
and exercise timeout handling.

### Run

```bash
go run ./cmd/testserver
```

You can display the available options with:

```bash
go run ./cmd/testserver --help
```

### Flags

| Flag       | Description                 | Default |
|------------| --------------------------- |---------|
| `--status` | Response status code        | `204`   |
| `--delay`  | Artificial response delay   | `0s`    |

---

## Docker

Build the CLI image:

```bash
docker build -t notifier .
```