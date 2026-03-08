.PHONY: dev
dev:
	@fuser -k 8080/tcp || true && go run . --dev=true

.PHONY: run
run:
	@fuser -k 8080/tcp || true && go run .
