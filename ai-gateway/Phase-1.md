Think of it as building a simplified version of the gateway sitting in front of OpenAI, Anthropic, or a self-hosted inference cluster.

For Phase 1, however, we won't connect to any real AI model. The gateway will simply proxy requests to mocked backend services. That lets us focus entirely on writing excellent Go.

Project
AI Gateway

A high-performance HTTP gateway responsible for:

Receiving client requests
Authenticating users
Applying middleware
Routing requests
Forwarding requests to upstream services
Logging
Collecting metrics
Returning responses

No Kubernetes.

No databases.

No AI.

No Redis.

No Docker.

Just one Go binary.

Architecture
Client

                   │
                   ▼

        ┌────────────────────┐
        │    HTTP Server      │
        └────────────────────┘
                   │
                   ▼
        ┌────────────────────┐
        │    Middleware       │
        └────────────────────┘
           │
           ├── Logging
           ├── Authentication
           ├── Request ID
           ├── Panic Recovery
           ├── Timeout
           └── Rate Limiting (in-memory)

                   │
                   ▼
        ┌────────────────────┐
        │      Router         │
        └────────────────────┘
                   │
         ┌─────────┴──────────┐
         ▼                    ▼

Mock Model A Mock Model B

Everything lives inside one process.

Learning Goals

Every feature should teach a Go concept.

Feature Go Concepts
HTTP server net/http
Router Interfaces
Middleware Higher-order functions
Context context.Context
Logging Structured types
Authentication Packages
Reverse Proxy net/http/httputil
Graceful shutdown Goroutines
Timeouts Context cancellation
Config Environment variables
Testing testing package
Benchmarking Benchmark tests
Functional Requirements

1. HTTP Server

Endpoints:

GET /health

GET /version

POST /v1/chat/completions

POST /v1/embeddings

POST /v1/audio

GET /metrics

The AI endpoints won't perform inference yet—they'll forward requests to mock services.

2. Reverse Proxy

The gateway should proxy requests to different upstream servers.

Example:

POST /v1/chat/completions

↓

Gateway

↓

http://localhost:9001/chat

or

↓

http://localhost:9002/chat

depending on routing rules.

This introduces one of the core responsibilities of modern API gateways.

3. Configuration

Everything configurable through environment variables.

Example:

PORT=8080

MODEL_A_URL=http://localhost:9001

MODEL_B_URL=http://localhost:9002

JWT_SECRET=...

REQUEST_TIMEOUT=30

LOG_LEVEL=info

No YAML files yet.

4. Structured Logging

Every request should produce logs like:

{
"request_id":"abc123",
"method":"POST",
"path":"/v1/chat/completions",
"status":200,
"latency_ms":37,
"client_ip":"127.0.0.1"
}

You'll learn:

log/slog
custom handlers
request-scoped logging 5. Request IDs

Every request receives:

X-Request-ID

If supplied by the client:

Reuse it.

Otherwise:

Generate one.

Pass it to upstream services.

This teaches Context usage.

6. Authentication

Simple Bearer token.

Authorization: Bearer abc123

Validate:

Missing token
Invalid token
Valid token

We'll deliberately keep JWTs for a later iteration to focus on gateway mechanics.

7. Panic Recovery

No panic should crash the server.

Instead:

{
"error":"internal server error"
}

with HTTP 500.

8. Request Timeout

Every request gets:

30 second timeout

If exceeded:

Return

504 Gateway Timeout

using context.Context.

9. Rate Limiter

Simple in-memory limiter.

Example:

100 requests/minute

per API key

Not distributed.

No Redis.

Just Go maps plus mutexes.

You'll learn:

sync.Mutex
time
maps
concurrency 10. Router

Routing rules:

chat

↓

Model A

embeddings

↓

Model B

Later we'll evolve this into dynamic routing.

11. Health Endpoint
    GET /health

Returns

{
"status":"ok"
} 12. Metrics

We'll expose simple metrics ourselves first, for example:

{
"requests":1203,
"errors":17,
"average_latency":23
}

Later phases can replace this with Prometheus without changing much of the surrounding architecture.

Internal Project Structure

I recommend a layout that is idiomatic in Go and scales naturally:

ai-gateway/
├── cmd/
│ └── gateway/
│ └── main.go
│
├── internal/
│ ├── auth/
│ ├── config/
│ ├── gateway/
│ ├── handlers/
│ ├── logging/
│ ├── middleware/
│ ├── metrics/
│ ├── proxy/
│ ├── router/
│ └── requestid/
│
├── pkg/
│ └── response/
│
├── test/
│
├── go.mod
└── README.md

This separates the executable (cmd) from implementation details (internal) and any reusable packages (pkg).

What we're deliberately not building

To keep the project focused, we will postpone:

❌ Kubernetes
❌ Docker
❌ Redis
❌ PostgreSQL
❌ Vector databases
❌ JWT validation
❌ OAuth
❌ TLS termination
❌ Load balancing across multiple instances
❌ Circuit breakers
❌ Retries
❌ Caching
❌ AI inference
❌ Streaming responses
❌ OpenTelemetry
❌ Prometheus
❌ Grafana

Each of these belongs in a later phase, where it naturally builds on the foundation you're creating now.

One additional goal

Rather than simply "getting it to work," treat this as if it were a production codebase under review by senior infrastructure engineers.

For every package, ask:

Is there a single, well-defined responsibility?
Is the public API minimal and clear?
Does it rely on interfaces where appropriate?
Can it be tested independently?
Does it handle errors explicitly instead of hiding them?
Would I be comfortable maintaining this code a year from now?

That mindset is what separates experienced application developers from engineers who build infrastructure platforms. This project should showcase not just that you know Go, but that you understand how to structure and evolve a production-quality service.
