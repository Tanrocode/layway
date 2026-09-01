# layway

A lightweight AI gateway written in Go. Built with Redis-backed semantic caching, OTel distributed tracing, and multi-provider access. 

## Routes

| Route | Behavior |
|---|---|
| `POST /v1/chat/completions` | Tries all configured providers in order, with fallback |
| `POST /openai/v1/chat/completions` | Forces OpenAI only |
| `POST /anthropic/v1/messages` | Forces Anthropic only |

All routes accept the same unified request shape:

```json
{
  "model": "gpt-4o-mini",
  "messages": [{ "role": "user", "content": "hello" }]
}
```

## Running it

```
export OPENAI_API_KEY=sk-...       # or put these in a .env file
export ANTHROPIC_API_KEY=sk-ant-...
go run ./cmd/gateway
```

```
curl localhost:8080/v1/chat/completions -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

## Architecture

```
cmd/gateway/          entrypoint — env/config loading, route wiring
internal/provider/     Provider interface + OpenAI/Anthropic implementations
internal/schema/       unified request/response types shared across providers
internal/gateway/      HTTP handler, retry logic, rate limiting, logging middleware
```

*No performance numbers are published here yet — they'll be added once measured against a real benchmark, not estimated.*
