include common.mk
include tools.mk

LDFLAGS += -X "$(MODULE)/version.Version=$(VERSION)" -X "$(MODULE)/version.CommitSHA=$(VERSION_HASH)"

## Build:

.PHONY: build
build: | build-frontend build-backend ## Build binary

.PHONY: dev
dev: | build-frontend build-backend-dev ## Build binary

.PHONY: build-frontend
build-frontend: ## Build frontend
	$Q cd frontend && yarn --frozen-lockfile && yarn build

.PHONY: build-backend-dev
build-backend-dev: ## Build backend-dev
	$Q $(go) build -ldflags '$(LDFLAGS)' -o .

.PHONY: build-backend
build-backend: ## Build backend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o .

.PHONY: test
test: | test-frontend test-backend ## Run all tests

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run typecheck

.PHONY: test-backend
test-backend: ## Run backend tests
	$Q $(go) test -v ./...

.PHONY: lint
lint: lint-frontend lint-backend ## Run all linters

.PHONY: lint-frontend
lint-frontend: ## Run frontend linters
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run lint

.PHONY: lint-backend
lint-backend: | $(golangci-lint) ## Run backend linters
	$Q $(golangci-lint) run -v

.PHONY: lint-commits
lint-commits: $(commitlint) ## Run commit linters
	$Q ./scripts/commitlint.sh

fmt: $(goimports) ## Format source files
	$Q $(goimports) -local $(MODULE) -w $$(find . -type f -name '*.go' -not -path "./vendor/*")

clean: clean-tools ## Clean

## Release:

.PHONY: bump-version
bump-version: $(standard-version) ## Bump app version
	$Q ./scripts/bump_version.sh

## Help:
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target> [options]${RESET}'
	@echo ''
	@echo 'Options:'
	@$(call global_option, "V [0|1]", "enable verbose mode (default:0)")
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
