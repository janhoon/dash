.PHONY: help backend frontend

help:
	@printf "Available targets:\n"
	@printf "  make backend   Start Go backend (hot reload with air if installed)\n"
	@printf "  make frontend  Start Vite frontend dev server\n"

backend:
	@if command -v air >/dev/null 2>&1; then \
		cd backend && air; \
	else \
		printf "air is not installed. Falling back to go run ./cmd/api\n"; \
		printf "Install air for hot reload: go install github.com/air-verse/air@latest\n"; \
		cd backend && go run ./cmd/api; \
	fi

frontend:
	@cd frontend && npm run dev
