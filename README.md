# zengine

Monorepo for the zengine stack.

## Repository Structure

```text
.
|-- .github/                 # CI/CD workflows
|-- hack/                    # helper scripts
|-- laravel/                 # Laravel backend (route /api)
|-- mwz/                     # MediaWiki custom resources
|   |-- extensions/
|   |   `-- ZetaExtension/   # custom wiki extension
|   `-- skins/
|       `-- ZetaSkin/        # custom wiki skin
|           |-- dist/        # skin Svelte build output
|           |-- svelte/      # skin Svelte project
|           `-- templates/   # skin Mustache templates
|-- svelte/                  # main Svelte project
|   `-- dist/                # main Svelte build output (route /)
`-- w/                       # MediaWiki core source (route /w & /wiki)
```
