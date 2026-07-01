# AI Gateway (Phase 1)

A single-binary HTTP gateway that authenticates, rate-limits, and proxies AI-style API requests to mock upstream services.

## Prerequisites

- Go 1.22+

## Local development

Run the mock upstreams and gateway in separate terminals.

**Terminal 1 — Mock Model A (chat):**

```bash
go run ./cmd/mock-model-a
```

Listens on `:9001` by default. Override with `PORT=9001`.

**Terminal 2 — Mock Model B (embeddings, audio):**

```bash
go run ./cmd/mock-model-b
```

Listens on `:9002` by default. Override with `PORT=9002`.

**Terminal 3 — Gateway:**

```bash
export API_TOKEN=dev-token
go run ./cmd/gateway
```

Listens on `:8080` by default.

## Example requests

Replace `dev-token` with your `API_TOKEN` value.

```bash
# Health (no auth)
curl http://localhost:8080/health

# Version (no auth)
curl http://localhost:8080/version

# Metrics (no auth)
curl http://localhost:8080/metrics

# Chat → mock-model-a
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer dev-token" \
  -H "Content-Type: application/json" \
  -d '{}'

# Embeddings → mock-model-b
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Authorization: Bearer dev-token" \
  -H "Content-Type: application/json" \
  -d '{}'

# Audio → mock-model-b
curl -X POST http://localhost:8080/v1/audio \
  -H "Authorization: Bearer dev-token" \
  -H "Content-Type: application/json" \
  -d '{}'
```

You can also hit the mocks directly:

```bash
curl -X POST http://localhost:9001/chat
curl -X POST http://localhost:9002/embeddings
curl -X POST http://localhost:9002/audio
```

## Configuration

| Env var           | Default                 | Purpose                          |
| ----------------- | ----------------------- | -------------------------------- |
| `PORT`            | `8080`                  | Gateway listen port              |
| `MODEL_A_URL`     | `http://localhost:9001` | Chat upstream                    |
| `MODEL_B_URL`     | `http://localhost:9002` | Embeddings/audio upstream        |
| `API_TOKEN`       | (required)              | Valid bearer token               |
| `REQUEST_TIMEOUT` | `30s`                   | Per-request deadline             |
| `RATE_LIMIT`      | `100`                   | Requests per minute per API key  |
| `LOG_LEVEL`       | `info`                  | slog level                       |
