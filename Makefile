.DEFAULT_GOAL := test-up
DOCKER_COMPOSE ?= docker compose
PSQL ?= docker compose exec -T db psql -U postgres

.PHONY: test-up down migrate-test test-only help ensure-test-db migrate-test-db

help:
	@printf "%s\n" \
	"Available targets:" \
	"  test-up      - migrations + integration tests + start app (default)" \
	"  test-only    - migrations + integration tests (no app)" \
	"  migrate-test - only migrations for pr_test_db" \
	"  down         - stop and remove containers" \
	"  help         - show this help"

# Internal helper: wait for primary pr_db readiness
wait-db:
	@echo "[wait-db] Waiting for PostgreSQL to become ready..."
	@retries=30; \
	while ! $(PSQL) -d pr_db -c '\q' >/dev/null 2>&1; do \
		retries=$$((retries-1)); \
		[ $$retries -le 0 ] && echo "[wait-db] Timeout waiting for db" && exit 1; \
		sleep 1; \
	done; \
	echo "[wait-db] DB is ready."

# Internal helper: ensure pr_test_db exists
ensure-test-db: wait-db
	@echo "[ensure-test-db] Ensuring pr_test_db exists..."
	@$(PSQL) -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='pr_test_db'" | grep -q 1 || $(PSQL) -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE pr_test_db WITH OWNER postgres TEMPLATE template0"
	@echo "[ensure-test-db] Done."

# Internal helper: apply migrations to pr_test_db if pull_requests table missing
migrate-test-db: ensure-test-db
	@echo "[migrate-test-db] Checking schema..."
	@$(PSQL) -d pr_test_db -tAc "SELECT 1 FROM information_schema.tables WHERE table_name='pull_requests'" | grep -q 1 || \
	( echo "[migrate-test-db] Applying migrations from /docker-entrypoint-initdb.d..."; for f in $$(docker compose exec -T db sh -c 'ls /docker-entrypoint-initdb.d/*.sql 2>/dev/null'); do echo "[migrate-test-db] Applying $$f"; docker compose exec -T db psql -U postgres -d pr_test_db -v ON_ERROR_STOP=1 -f $$f; done; echo "[migrate-test-db] Migrations applied." )
	@echo "[migrate-test-db] Done."

migrate-test:
	@echo "[migrate-test] Starting db container..."
	$(DOCKER_COMPOSE) up -d db
	@$(MAKE) migrate-test-db
	@echo "[migrate-test] Completed."

test-only:
	@echo "[test-only] Starting db container..."
	$(DOCKER_COMPOSE) up -d db
	@$(MAKE) migrate-test-db
	@echo "[test-only] Running integration tests..."
	@if ! $(DOCKER_COMPOSE) --profile test run --rm integration-tests; then \
		echo "[test-only] Integration tests failed."; \
		exit 1; \
	fi
	@echo "[test-only] Integration tests passed."

test-up:
	@echo "[test-up] Starting db container..."
	$(DOCKER_COMPOSE) up -d db
	@$(MAKE) migrate-test-db
	@echo "[test-up] Running integration tests..."
	@if ! $(DOCKER_COMPOSE) --profile test run --rm integration-tests; then \
		echo "[test-up] Integration tests failed. Stopping stack..."; \
		$(DOCKER_COMPOSE) down; \
		exit 1; \
	fi
	@echo "[test-up] Integration tests passed. Starting app..."
	$(DOCKER_COMPOSE) up -d app
	@echo "[test-up] App is running."

down:
	$(DOCKER_COMPOSE) down -v
