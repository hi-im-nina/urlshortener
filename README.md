# ⭐ 🔗 URL Shortener 🔗 ⭐

A simple URL shortening service written in Go. Supports both in-memory storage (for local development) and Redis (for persistence).

## 🏗️ Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- Redis (optional — the app falls back to in-memory storage if Redis isn't running)

## 🏃🏻‍♀️ Running the server

```bash
go run .
```

The server starts on `http://localhost:8080`. You'll see one of these in your terminal:

```
Connected to Redis at localhost:6379        # Redis is running
Redis unavailable — falling back to in-memory store  # No Redis, that's fine
```

## 🗃️ Running with Redis (optional)

If you have [Docker](https://www.docker.com/) installed, the easiest way to start Redis is:

```bash
docker run -p 6379:6379 redis
```

Then in a separate terminal, run `go run .` and you'll see it connect.

## API Documentation

### Shorten a URL

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://hi-im-nina.github.io"}'
```

Response:

```json
{
  "short_url": "http://localhost:8080/aB3xZ9",
  "original_url": "https://hi-im-nina.github.io/"
}
```

### Redirect to the original URL

Open `http://localhost:8080/<code>` in your browser, or:

```bash
curl -L http://localhost:8080/aB3xZ9
```

### View stats for a short URL

```bash
curl http://localhost:8080/stats/aB3xZ9
```

Response:

```json
{
  "short_url": "http://localhost:8080/aB3xZ9",
  "original_url": "https://hi-im-nina.github.io/",
  "clicks": 3,
  "created_at": "2026-04-09 10:00:00"
}
```

## 🧪 Running tests

```bash
go test ./...
```
