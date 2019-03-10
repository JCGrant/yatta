package main

import (
	"log"
	"os"
	"strconv"

	"github.com/JCGrant/yatta/notifiers"
	"github.com/JCGrant/yatta/server"
	"github.com/JCGrant/yatta/todos"
	"github.com/pkg/errors"
)

func main() {
	notifier, err := notifiers.NewGmailNotifier(
		"james@jcgrant.com",
		"./secrets/credentials.json",
		"./secrets/token.json")
	if err != nil {
		log.Fatal(err)
	}
	todoManager, err := todos.NewManager("./secrets/todos.json")
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "getting port failed"))
	}
	s := server.New(port, todoManager, notifier)
	log.Fatal(s.Start())
}
