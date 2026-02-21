.PHONY:

build:
	go build -o ./.bin/bot cmd/bot/main.go

run:
	./.bin/bot

enterd:
	docker exec -it kachka_bot-db-1 psql -U dmytro -d kachka_bot
