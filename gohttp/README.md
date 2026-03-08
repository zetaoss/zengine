# gohttp

Go HTTP server for zengine front routing.

## Modes

- `dev`
  - Proxies frontend requests to Vite (`http://127.0.0.1:5173`, fixed)
  - Supports WebSocket upgrade for Vite HMR through proxy
  - Injects runtime config script into HTML responses
- `prod`
  - Serves static files from `/app/svelte/dist` (fixed)
  - If a requested file exists, serve it; otherwise serve injected `index.html` for routes and return `404` for static asset paths
  - Injects runtime config script into served `index.html`

## Run

```bash
cd /app/gohttp
go run . --dev=true
```

```bash
cd /app/gohttp
go run .
```

Example:

```bash
go run . --dev=true
```

## Runtime config injection

Runtime config is injected as:

```html
<script>window.ZCONF={...};</script>
```

Injected values are built from:
- `AD_CLIENT`
- `AD_SLOTS` (comma-separated, e.g. `foo,bar`)
- `AVATAR_BASE_URL`
- `GA_MEASUREMENT_ID`
- request header `X-Policy`
