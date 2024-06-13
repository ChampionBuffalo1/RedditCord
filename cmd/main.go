package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	handler "github.com/ChampionBuffalo1/redditcord/internal/handlers"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Failed to load .env file")
	}
	TOKEN := os.Getenv("TOKEN")
	if TOKEN == "" {
		log.Fatalln("Missing Bot Token!")
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", TOKEN))
	if err != nil {
		log.Println(err)
	}

	handler.ImplementHandlers(discord)

	if err := discord.Open(); err != nil {
		log.Printf("Got error while establishing connection with discord: %s", err)
	}
	defer discord.Close()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	log.Println("Gracefully shutdown bot.")
}
