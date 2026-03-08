# Repository Structure

This repo is a monorepo that contains multiple stacks.

```text
.
|-- .github/                                  # CI/CD workflows
|-- hack/                                     # helper scripts
|-- gohttp/                                   # Go HTTP server for frontend routing (port 8080)
|-- laravel/                                  # Laravel backend (route /api)
|-- mwz/                                      # MediaWiki custom resources
|   |-- extensions/
|   |   `-- ZetaExtension/                    # ZetaExtension
|   `-- skins/
|       `-- ZetaSkin/                         # ZetaSkin
|           |-- dist/                         # - skin Svelte build output
|           |-- svelte/                       # - skin Svelte: plain Svelte + Vite project
|           |   `-- src/lib/shared (symlink)  # - points to main shared source
|           `-- templates/                    # - skin Mustache templates
|-- svelte/                                   # main Svelte: SvelteKit project
|   |-- src/lib/shared                        # - main shared source
|   `-- dist/                                 # - main Svelte build output (route /)
`-- w/                                        # MediaWiki core source (route /w & /wiki)
```
