package main

import (
	"Raphael/core/Handler"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	client, err := discordgo.New("Bot " + os.Getenv("DISOCRD_TOKEN"))
	if err != nil {
		slog.Error("Error creating Discord session", err)
		return
	}

	client.AddHandler(Handler.MessageCreate)
	loadInteractionCommand(client)
	client.AddHandler(Handler.InteractionCreate)
	client.Identify.Intents = discordgo.IntentGuildMessages

	err = client.Open()
	if err != nil {
		slog.Error("Error opening connection", err)
		return
	}

	fmt.Println("Raphael is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	client.Close()
}
