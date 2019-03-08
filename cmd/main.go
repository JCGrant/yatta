package main

import (
	"log"

	"github.com/JCGrant/yatta/notifiers"
)

func main() {
	notifier, err := notifiers.NewGmailNotifier(
		"james@jcgrant.com",
		"./secrets/credentials.json",
		"./secrets/token.json")
	if err != nil {
		log.Fatal(err)
	}
	notifier.Notify("hello")
}
