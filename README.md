# QuotesMCP

MCP server that exposes a `get_quote` tool powered by [API Ninjas](https://api-ninjas.com).

## Prerequisites
- Go 1.25+
- API Ninjas account + API key

## Setup

1. Clone the repo
2. Copy `.env.example` to `.env` and fill in your values:
   - `API_NINJAS_KEY` — required
   - `BASE_URL` — required, public base URL of the server (e.g. `https://your-domain.example.com`)
   - `PORT` — defaults to `8080` if unset (project uses `7211`)
   - `DEPLOY_TARGET` — only needed for deployment

## Build

```bash
./build.sh
```

Produces a `quotes-mcp` binary (symbols stripped for smaller size).

## Run

```bash
./quotes-mcp
```

Health check: `GET /health` → `ok`

## MCP Tool: `get_quote`

| Parameter  | Type   | Required | Description                                              |
|------------|--------|----------|----------------------------------------------------------|
| `category` | string | No       | Filter by category (e.g. `success`, `wisdom`, `love`)   |

Returns: `"[quote]" — [author], [work]`

## Deploy

```bash
./build_and_deploy.sh
```

Builds the binary and SCPs it to `DEPLOY_TARGET` from `.env`.
