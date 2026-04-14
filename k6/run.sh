#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"

ENV_NAME="${1:?Usage: ./run.sh <local|vps> <script> [k6 options]}"
shift

if [[ "$ENV_NAME" != "local" && "$ENV_NAME" != "vps" ]]; then
  echo "Error: unknown environment '$ENV_NAME'. Use 'local' or 'vps'." >&2
  exit 1
fi

ENV_FILE=".env.k6.${ENV_NAME}"
if [ ! -f "$ENV_FILE" ]; then
  echo "Error: $ENV_FILE not found. Run: cp .env.k6.${ENV_NAME}.example $ENV_FILE" >&2
  exit 1
fi

set -a; source "$ENV_FILE"; set +a
k6 run "$@"
