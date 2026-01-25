.PHONY: local cloud down logs ps rebuild-local rebuild-cloud

local:
	set -a && source .env.local && set +a && docker compose --profile local up -d --build

cloud:
	set -a && source .env.cloud && set +a && docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

rebuild-local:
	docker compose down
	set -a && source .env.local && set +a && docker compose --profile local up -d --build

rebuild-cloud:
	docker compose down
	set -a && source .env.cloud && set +a && docker compose up -d --build
