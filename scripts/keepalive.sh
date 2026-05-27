#!/usr/bin/env bash
set -euo pipefail

URL="${KEEPALIVE_URL:-}"
INTERVAL_SECONDS="${KEEPALIVE_INTERVAL_SECONDS:-600}"

if [ -z "$URL" ]; then
  echo "KEEPALIVE_URL is required. Example: KEEPALIVE_URL=https://your-service.onrender.com/health ./scripts/keepalive.sh"
  exit 1
fi

echo "Starting keep-alive pings to $URL every $INTERVAL_SECONDS seconds"

while true; do
  timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  status="$(curl -L -s -o /dev/null -w "%{http_code}" --max-time 20 "$URL" || true)"
  echo "[$timestamp] GET $URL -> HTTP $status"
  sleep "$INTERVAL_SECONDS"
done
