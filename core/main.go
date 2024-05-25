package main

import (
	"Raphael/core/Handler"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := discordgo.New("Bot " + os.Getenv("DISOCRD_TOKEN"))
	if err != nil {
		log.Fatal("Error creating Discord session", err)
		return
	}

	client.AddHandler(Handler.MessageCreate)
	client.Identify.Intents = discordgo.IntentGuildMessages

	err = client.Open()
	if err != nil {
		log.Fatal("Error opening connection", err)
		return
	}

	fmt.Println("Raphael is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	client.Close()
}
