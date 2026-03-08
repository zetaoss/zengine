# gohttp

Go HTTP server for zengine front routing.

## Modes

- `dev`
  - Proxies frontend requests to Vite (`http://127.0.0.1:5173`, fixed)
  - Supports WebSocket upgrade for Vite HMR through proxy
  - Injects runtime config script into HTML responses
- `prod`
  - Serves static files from `/app/svelte/dist` (fixed)
  - Returns `404` when requested file is missing
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
- `ADSENSE_CLIENT`
- `ADSENSE_SLOTS` (comma-separated, e.g. `slotTop,slotBottom`)
- `AVATAR_BASE_URL`
- `GA_MEASUREMENT_ID`
- request header `X-Policy`
