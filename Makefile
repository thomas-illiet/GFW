.PHONY: build test coverage coverage-check race vet fmt-check comment-check size-check openapi-check helm-check compose-check version-check verify frontend-install frontend-build run migrate demo-seed image image-smoke chart-package acceptance release-check

GOCACHE ?= /private/tmp/great-firewall-go-cache
VERSION ?= 0.1.0
IMAGE ?= gfw
COMMIT := $(or $(COMMIT),$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown))
BUILD_DATE := $(or $(BUILD_DATE),$(shell date -u +%Y-%m-%dT%H:%M:%SZ))

frontend-install:
	cd frontend && npm ci

frontend-build:
	cd frontend && npm run generate

build: frontend-build
	GOCACHE=$(GOCACHE) go build -o bin/great-firewall ./cmd/gfw

test:
	GOCACHE=$(GOCACHE) go test ./...
	cd frontend && npm run test

coverage:
	mkdir -p coverage
	GOCACHE=$(GOCACHE) go test ./internal/... -coverprofile=coverage/backend.out
	GOCACHE=$(GOCACHE) go tool cover -func=coverage/backend.out | tee coverage/backend.txt
	GOCACHE=$(GOCACHE) go tool cover -html=coverage/backend.out -o coverage/backend.html

coverage-check: coverage
	awk '/^total:/ { gsub("%", "", $$3); if ($$3 + 0 < 80) { print "backend coverage " $$3 "% is below 80%"; exit 1 } }' coverage/backend.txt

race:
	GOCACHE=$(GOCACHE) go test -race ./internal/...

vet:
	GOCACHE=$(GOCACHE) go vet ./...

fmt-check:
	test -z "$$(gofmt -l cmd internal)"

comment-check:
	@awk 'BEGIN { bad=0 } /^func / || /^func \(/ { if (prev !~ /^\/\//) { print FILENAME ":" FNR ": undocumented function"; bad=1 } } { prev=$$0 } END { exit bad }' $$(find cmd internal -name '*.go' ! -name '*_test.go')

size-check:
	@find cmd internal -name '*.go' ! -name '*_test.go' -exec awk 'FNR==351 { print FILENAME ": exceeds 350 lines"; bad=1 } END { if (bad) exit 1 }' {} +

openapi-check:
	GOCACHE=$(GOCACHE) go run ./scripts/check_openapi.go

helm-check:
	helm lint charts/gfw
	helm template gfw charts/gfw --set secrets.databaseURL=postgres://example --set secrets.oidcClientSecret=secret --set secrets.sessionSecret=12345678901234567890123456789012 >/dev/null
	helm template gfw charts/gfw --set secrets.existingSecret=gfw-runtime >/dev/null
	@test -z "$$(helm template gfw charts/gfw --set secrets.existingSecret=gfw-runtime | grep -Ei 'redis|opa|component: worker')"

compose-check:
	docker compose config >/dev/null

version-check:
	@test "$$(awk '/^version:/ {print $$2; exit}' charts/gfw/Chart.yaml)" = "$(VERSION)"
	@test "$$(awk '/^appVersion:/ {gsub(/\"/, "", $$2); print $$2; exit}' charts/gfw/Chart.yaml)" = "$(VERSION)"
	@test "$$(awk '/^  version:/ {print $$2; exit}' openapi.yaml)" = "1.0.0"

verify: fmt-check vet comment-check size-check coverage-check race openapi-check helm-check compose-check version-check
	cd frontend && npm run typecheck
	cd frontend && npm run lint:check
	cd frontend && npm run format:check
	cd frontend && npm run test
	cd frontend && npm run generate

run:
	go run ./cmd/gfw serve --dev-auth

migrate:
	go run ./cmd/gfw migrate up --dev-auth

demo-seed:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U gfw -d gfw < scripts/seed_demo.sql

image:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg BUILD_DATE=$(BUILD_DATE) -t $(IMAGE):$(VERSION) .

image-smoke: image
	docker run --rm $(IMAGE):$(VERSION) version | grep -F "$(VERSION) (commit $(COMMIT), built $(BUILD_DATE))"

chart-package: version-check
	mkdir -p dist
	helm package charts/gfw --destination dist

acceptance:
	./scripts/release_acceptance.sh

release-check: verify acceptance image-smoke chart-package
