# go-nexus

A WebSocket broadcast server based on Redis pub/sub.

Clients connect via WebSocket, send messages, and every message is broadcast to all other connected clients through Redis. Useful for real-time chat, live dashboards, or any app needing multi-instance pub/sub.

## Endpoints

| Path | Description |
|---|---|
| `GET /ws` | WebSocket upgrade — connect and send/receive messages |
| `GET /ping` | Health check — returns `{"message": "pinged!"}` |
| `GET /debug` | Debug info — returns connected client count |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `REDIS_URL` | `redis://localhost:6379` | Full Redis connection URL |
| `REDIS_ADDR` | `localhost:6379` | Fallback if `REDIS_URL` fails to parse |
| `REDIS_CHANNEL` | `go-nexus` | Redis pub/sub channel name |
| `PORT` | `8080` | HTTP server port |
| `SENTRY_DSN` | — | Sentry DSN for error monitoring (optional) |

## Local Development

```bash
# Redis must be running on localhost:6379
go run .
```

A `.env` file is provided with defaults — `godotenv` loads it automatically.

## Docker

```bash
docker build -t go-nexus .
docker run -e REDIS_URL=redis://host.docker.internal:6379 -p 8080:8080 go-nexus
```

## Sentry Error Monitoring

Panics in any request handler are automatically captured and sent to Sentry. The server also captures Redis pub/sub failures.

1. Sign up at [sentry.io](https://sentry.io) and create a Go project
2. Copy your DSN (starts with `https://...@...`)
3. Set it as `SENTRY_DSN` in your environment

Works locally via `.env` or on Render via the dashboard.

## Deploy to Render

1. Push this repo to GitHub
2. Update `repo` in `render.yaml` to your GitHub URL
3. Connect your GitHub repo to Render and it auto-detects the Blueprint
4. Set `SENTRY_DSN` in Render's dashboard

The `render.yaml` provisions a free web service + free Redis instance and wires `REDIS_URL` automatically.
