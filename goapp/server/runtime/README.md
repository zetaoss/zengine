# Runtime Package

This package is responsible for configuring core components based on the application's execution environment (dev/prod).

## Structure

- `common/`: Contains middleware and dependency injection logic used across all environments.
- `dev/`: Development-specific components. Handles integration with the Vite development server (Reverse Proxy) and injecting development-specific scripts.
- `prod/`: Production-specific components. Handles serving built static files, caching `index.html`, and SPA fallback processing.

## Key Responsibilities

1. **Dependency Injection**: Dynamically injects client-side configuration (`window.ZCONF`) into the HTML.
2. **Middleware Configuration**: Configures environment-specific middlewares (e.g., access logs).
3. **Root Handler**: Manages routing and fallback logic between frontend static resources and the API server.

## Design Rationale: Config Injection Strategy

We use simple string replacement targeting `</title>` to inject `window.ZCONF`. This avoids the overhead of a full HTML parser while remaining highly robust:

- **Reliable & Fast**: `</title>` is a mandatory fixed string that minifiers never strip, making `bytes.Replace` both safe and fast.
- **Execution Order**: Ensures config is loaded before any other scripts.

## Notes

- **AccessLogMiddleware**: While the core logging logic resides in `common/middleware.go`, each environment provides its own wrapper:
    - `dev`: Implements `http.Hijacker` support to handle upgraded connections (e.g., Vite HMR WebSockets).
    - `prod`: Uses a standard implementation focused on performance and reliability.
- The `index.html` in the `prod` environment is loaded into memory at server startup, so a server restart may be required if the file changes.
