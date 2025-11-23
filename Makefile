.DEFAULT_GOAL := test-up
DOCKER_COMPOSE ?= docker compose

.PHONY: test-up down migrate-test test-only help

help:
	@echo "Available targets:";
	@echo "  test-up      - migrations + integration tests + start app (default)";
	@echo "  test-only    - migrations + integration tests (no app)";
	@echo "  migrate-test - only migrations for pr_test_db";
	@echo "  down         - stop and remove containers";
	@echo "  help         - show this help"

migrate-test:
	@echo "[migrate-test] Starting DB..."
	$(DOCKER_COMPOSE) up -d db
	@echo "[migrate-test] Running migrations via test-db-migrator..."
	@if ! $(DOCKER_COMPOSE) run --rm test-db-migrator; then \
		echo "[migrate-test] Migration failed."; \
		exit 1; \
	fi
	@echo "[migrate-test] Migrations completed."

test-only:
	@echo "[test-only] Starting DB..."
	$(DOCKER_COMPOSE) up -d db
	@echo "[test-only] Running migrations via test-db-migrator..."
	@if ! $(DOCKER_COMPOSE) run --rm test-db-migrator; then \
		echo "[test-only] Migration failed."; \
		exit 1; \
	fi
	@echo "[test-only] Migrations completed. Running integration tests..."
	@if ! $(DOCKER_COMPOSE) run --rm integration-tests; then \
		echo "[test-only] Integration tests failed."; \
		exit 1; \
	fi
	@echo "[test-only] Integration tests passed."

test-up:
	@echo "[test-up] Starting DB..."
	$(DOCKER_COMPOSE) up -d db
	@echo "[test-up] Running migrations via test-db-migrator..."
	@if ! $(DOCKER_COMPOSE) run --rm test-db-migrator; then \
		echo "[test-up] Migration failed. Stopping stack..."; \
		$(DOCKER_COMPOSE) down; \
		exit 1; \
	fi
	@echo "[test-up] Migrations completed. Running integration tests..."
	@if ! $(DOCKER_COMPOSE) run --rm integration-tests; then \
		echo "[test-up] Integration tests failed. Stopping stack..."; \
		$(DOCKER_COMPOSE) down; \
		exit 1; \
	fi
	@echo "[test-up] Integration tests passed. Starting app..."
	$(DOCKER_COMPOSE) up -d app
	@echo "[test-up] App is running."

down:
	$(DOCKER_COMPOSE) down