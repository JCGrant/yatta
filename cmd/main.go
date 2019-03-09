package main

import (
	"log"

	"github.com/JCGrant/yatta/notifiers"
	"github.com/JCGrant/yatta/server"
	"github.com/JCGrant/yatta/todos"
)

func main() {
	notifier, err := notifiers.NewGmailNotifier(
		"james@jcgrant.com",
		"./secrets/credentials.json",
		"./secrets/token.json")
	if err != nil {
		log.Fatal(err)
	}
	todoManager := todos.NewManager()
	s := server.New(8080, todoManager, notifier)
	log.Fatal(s.Start())
}
