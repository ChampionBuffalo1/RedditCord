package interactions

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(session *discordgo.Session) {
	MinValue, MaxValue := 1.0, 21.0
	command := &discordgo.ApplicationCommand{
		Name:        "reddit",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Gets recent posts from any public subreddit",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionType(3),
				Name:         "subreddit",
				Description:  "Name of the subreddit to fetch",
				Required:     false,
				MinValue:     &MinValue,
				MaxValue:     MaxValue,
				ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
				Autocomplete: true,
			}},
	}
	session.ApplicationCommandCreate(os.Getenv("APP_ID"), "", command)
}
