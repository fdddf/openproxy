SHELL := /bin/bash

.PHONY: docker-image
docker-image: ## Build and push Docker image via build.sh
	./build.sh

.PHONY: deploy
deploy: ## Build and deploy to Kubernetes
	bash ./build.sh
	kubectl apply -f deploy_k8s.yaml
	echo "Deploy done"

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'