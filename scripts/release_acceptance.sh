#!/bin/sh
set -eu

project="gfw-release-$$"
export GFW_HTTP_PORT=18080
export GFW_POSTGRES_PORT=15432

cleanup() {
  docker compose -p "$project" down --volumes --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker compose -p "$project" build
docker compose -p "$project" up -d --wait postgres

docker compose -p "$project" run --rm migrate &
first=$!
docker compose -p "$project" run --rm migrate &
second=$!
wait "$first"
wait "$second"

docker compose -p "$project" run --rm migrate
docker compose -p "$project" run --rm migrate migrate version
docker compose -p "$project" run --rm migrate migrate status
docker compose -p "$project" up -d --wait great-firewall

curl --fail --silent --show-error "http://127.0.0.1:${GFW_HTTP_PORT}/health/ready" >/dev/null
curl --fail --silent --show-error "http://127.0.0.1:${GFW_HTTP_PORT}/auth/session" >/dev/null
