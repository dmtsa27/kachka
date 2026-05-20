include .env
.PHONY: build run enterd migrate dbsnap watchdb docker-up docker-down up reset swag

swag:
	swag init -g cmd/bot/main.go -o docs

build: swag
	go build -o ./.bin/bot cmd/bot/main.go

run: build
	./.bin/bot

enterd:
	docker exec -it kachka_bot-db-1 psql -U $(DB_USER) -d $(DB_NAME)

migrate:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" goose -dir ./migrations up

dbsnap:
	docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -x -c "SELECT now() AS ts; SELECT * FROM challenges ORDER BY id DESC LIMIT 5; SELECT * FROM challenge_bootstrap ORDER BY chat_id DESC LIMIT 5; SELECT * FROM chat_members ORDER BY last_seen_at DESC LIMIT 20; SELECT * FROM message_reactions ORDER BY updated_at DESC LIMIT 30; SELECT * FROM users ORDER BY telegram_id DESC LIMIT 20; SELECT * FROM workouts ORDER BY id DESC LIMIT 20; SELECT * FROM sessions ORDER BY id DESC LIMIT 20;"

watchdb:
	watch -n 2 $(MAKE) dbsnap

up: swag
	docker compose up -d db
	@echo "Waiting for database to be ready..."
	@until docker compose exec db pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
	$(MAKE) migrate
	docker compose up -d --build server

reset:
	docker compose down -v
	$(MAKE) up