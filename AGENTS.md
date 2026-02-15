# Repository Structure

This repo is a monorepo that contains multiple stacks.

```text
.
|-- .github/                 # CI/CD workflows
|-- hack/                    # helper scripts
|-- laravel/                 # Laravel backend (route /api)
|-- mwz/                     # MediaWiki custom resources
|   |-- extensions/
|   |   `-- ZetaExtension/   # ZetaExtension
|   `-- skins/
|       `-- ZetaSkin/        # ZetaSkin
|           |-- dist/        # - skin Svelte build output
|           |-- svelte/      # - skin Svelte project
|           `-- templates/   # - skin Mustache templates
|-- svelte/                  # main Svelte project
|   `-- dist/                # - main Svelte build output (route /)
`-- w/                       # MediaWiki core source (route /w & /wiki)
```
