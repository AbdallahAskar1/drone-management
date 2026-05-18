.PHONY: build test run up down migrate clean help \
        hurl hurl-auth hurl-enduser hurl-drone-happy hurl-drone-failed \
        hurl-drone-broken hurl-handoff hurl-heartbeat hurl-admin hurl-rbac \
        hurl-reset hurl-rebuild

BIN := server
CMD := ./cmd/server/main.go
export GOMODCACHE=$(shell pwd)/.go-cache
export GOCACHE=$(shell pwd)/.go-cache
export DOCKER_API_VERSION=1.44

build: ## Build the server binary
	mkdir -p .go-cache
	go build -o $(BIN) $(CMD)

test: ## Run all tests
	go test -v ./...

run: build ## Build and run the server locally
	./$(BIN)

up: ## Start docker-compose services
	docker-compose up -d

down: ## Stop docker-compose services
	docker-compose down

migrate: ## Run database migrations
	@echo "Running migrations..."
	# Add migration command here if using a specific tool

clean: ## Clean build artifacts
	rm -f $(BIN)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ── Hurl API tests ─────────────────────────────────────────────────────────────
HURL     := hurl
HURL_DIR := tests/hurl

hurl-reset: ## Truncate all DB tables for a clean test run (fast, no rebuild)
	docker exec drone-management-db-1 psql -U postgres -d drones \
	  -c "TRUNCATE principals, drones, orders, jobs, order_events CASCADE"

hurl-rebuild: ## Full teardown + image rebuild + fresh DB (use after code changes)
	docker-compose down -v
	docker-compose up -d --build
	@echo "Waiting for DB to be ready..."
	@sleep 5

hurl-auth: ## 01 – token issuance for all roles
	$(HURL) --test $(HURL_DIR)/01-auth.hurl

hurl-enduser: ## 02 – enduser order submit / get / withdraw
	$(HURL) --test $(HURL_DIR)/02-enduser.hurl

hurl-drone-happy: ## 03 – full happy-path delivery lifecycle
	$(HURL) --test $(HURL_DIR)/03-drone-happy-path.hurl

hurl-drone-failed: ## 04 – drone marks delivery as failed
	$(HURL) --test $(HURL_DIR)/04-drone-failed.hurl

hurl-drone-broken: ## 05 – drone self-reports broken (no goods, no handoff)
	$(HURL) --test $(HURL_DIR)/05-drone-broken-no-goods.hurl

hurl-handoff: ## 06 – broken-drone handoff: drone1 breaks, drone2 delivers
	$(HURL) --test $(HURL_DIR)/06-handoff.hurl

hurl-heartbeat: ## 07 – heartbeat updates location and ETA on the order
	$(HURL) --test $(HURL_DIR)/07-heartbeat.hurl

hurl-admin: ## 08 – admin endpoints + admin-fix does not cancel handoff job
	$(HURL) --test $(HURL_DIR)/08-admin.hurl

hurl-rbac: ## 09 – RBAC: wrong role → 403, no token → 401
	$(HURL) --test $(HURL_DIR)/09-rbac.hurl

hurl: hurl-reset ## Run all hurl API tests sequentially (auto-resets DB first)
	$(HURL) --test --jobs 1 $(HURL_DIR)/01-auth.hurl $(HURL_DIR)/02-enduser.hurl \
	  $(HURL_DIR)/03-drone-happy-path.hurl $(HURL_DIR)/04-drone-failed.hurl \
	  $(HURL_DIR)/05-drone-broken-no-goods.hurl $(HURL_DIR)/06-handoff.hurl \
	  $(HURL_DIR)/07-heartbeat.hurl $(HURL_DIR)/08-admin.hurl $(HURL_DIR)/09-rbac.hurl
