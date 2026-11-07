#!/usr/bin/env bash
# Run after make composer-lock; the container and its SQLite wiki are temporary.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
image=${MW_OAUTH_SMOKE_IMAGE:-zengine-base:composer-lock}
docker run --rm --entrypoint bash \
  -v "$root/hack/oauth-smoke-container.sh:/tmp/mw-oauth-flow.sh:ro" \
  -v "$root/hack/oauth-smoke.py:/tmp/mw-oauth-flow.py:ro" \
  -v "$root/mwz/settings/BaseExtensionSettings.php:/tmp/BaseExtensionSettings.php:ro" \
  "$image" /tmp/mw-oauth-flow.sh
