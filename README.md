# goping

A simple concurrent CLI tool for checking multiple URLs and reporting status code + latency.

## Features

- Concurrent URL checks using worker goroutines
- Configurable worker count
- Configurable HTTP method (`GET` or `HEAD`)
- Per-request timeout (5 seconds)
- Clear terminal output for success, warnings, and errors

## Requirements

- Go 1.20+

## Run

From the project root:

```bash
go run .
```

If no URLs are provided, the tool runs against a built-in default URL list.

## Usage

```bash
go run . [flags] <url1> <url2> ...
```

### Flags

- `-w` number of workers (default: `5`)
- `-m` HTTP method (default: `GET`)

## Examples

Check two URLs with default settings:

```bash
go run . https://example.com https://httpbin.org/status/404
```

Use 10 workers and `HEAD` requests:

```bash
go run . -w 10 -m HEAD https://google.com https://github.com
```

## Example Output

```text
✅ https://example.com             | 200 |  145ms
⚠️ https://httpbin.org/status/404 | 404 |  173ms
❌ https://invalid-url.test        | ERROR: Get "https://invalid-url.test": ...

--- Scan Complete ---
```

## Build Binary (optional)

```bash
go build -o goping .
./goping -w 8 https://example.com
```
