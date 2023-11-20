.DEFAULT_GOAL := help
.PHONY: dev up build build-no-cache createusers open-app submit help


dev: ## Setup dependencies needed for development
	go mod download
	cp -n .env.sample .env

up: ## Run all the services, web, Postgres, load-prices worker, load-securities worker
	docker compose up

build: ## Build the docker images used by services in the docker-compose.yml
	docker compose build

build-no-cache: ## Build the docker images, without the the docker build cache, used by the services in the docker-compose.yml
	docker compose build --no-cache

createusers: ## Create 2 standard users username=user1 and username=user2
	docker compose \
		exec \
		db \
		psql -U postgres -c "insert into users(user_id, email) values ('user1', 'user1@casestudy.com'), ('user2', 'user2@casestudy.com');"

open-app: ## Open the web app
	open http://localhost:8000

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
