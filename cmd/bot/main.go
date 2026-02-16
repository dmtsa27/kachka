package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	t := mustToken()
	fmt.Println(t)

}

func mustToken() string {

	token := flag.String("bot-token", "", "token for access to telegram bot")

	flag.Parse()

	if *token == "" {
		log.Fatal("invalid token")
	}

	return *token

}
