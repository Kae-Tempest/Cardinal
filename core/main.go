package main

import (
	"Raphael/core/handler"
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

	client, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		slog.Error("Error creating Discord session", err)
		return
	}

	err = client.Open()
	if err != nil {
		slog.Error("Error opening connection", err)
		return
	}

	client.AddHandler(handler.MessageCreate)
	loadInteractionCommand(client)
	client.AddHandler(handler.InteractionCreate)
	client.Identify.Intents = discordgo.IntentGuildMessages

	fmt.Println("Raphael is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	client.Close()
}
