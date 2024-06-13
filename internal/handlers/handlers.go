package handler

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	reddit "github.com/ChampionBuffalo1/redditcord/internal"
	"github.com/ChampionBuffalo1/redditcord/internal/interactions"
	"github.com/bwmarrin/discordgo"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Context struct {
	db *sql.DB
}

func InitDb() (*sql.DB, error) {
	db, err := sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", os.Getenv("TURSO_URL"), os.Getenv("TURSO_TOKEN")))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ImplementHandlers(session *discordgo.Session) {
	db, err := InitDb()
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}
	log.Println("Connected to SQLite")
	ctx := &Context{db: db}
	session.AddHandler(loginHandler)
	session.AddHandler(ctx.interactionHandler)
	interactions.RegisterCommands(session)
}

func loginHandler(session *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s\n", r.User.Username)
}

func (ctx *Context) interactionHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
		autocompleteHandler(session, interaction)
	} else if interaction.Type == discordgo.InteractionApplicationCommand {
		cmdName := interaction.ApplicationCommandData().Name
		if cmdName == "reddit" {
			ctx.handleRedditSlash(session, interaction)
		} else {
			log.Printf("Unknown Application Command: %s", cmdName)
		}
	} else {
		log.Printf("Unknown Interaction Type: %d", interaction.Type)
	}
}

func autocompleteHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()
	if data.Name == "reddit" {
		handleSubredditSearch(session, interaction)
	} else {
		log.Printf("Unhandled Autocomplete interaction for command %s", data.Name)
	}
}

func handleSubredditSearch(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()
	opts := data.Options[0]
	value := opts.StringValue()
	var choices []*discordgo.ApplicationCommandOptionChoice

	if value != "" && opts.Focused {
		ch := make(chan []*discordgo.ApplicationCommandOptionChoice)
		go reddit.GetSubreddits(value, ch)
		choices = <-ch
	}

	res := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(8),
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		}}
	session.InteractionRespond(interaction.Interaction, res)

}

func (ctx *Context) handleRedditSlash(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	cmd := interaction.ApplicationCommandData()
	subreddit := "memes" // default subreddit
	if len(cmd.Options) != 0 {
		subreddit = cmd.Options[0].StringValue()
	}

	ackRes := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(5),
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	ch := make(chan *reddit.RedditFetchResult)
	go reddit.FetchRedditPost(subreddit, ch)

	session.InteractionRespond(interaction.Interaction, ackRes)
	data := <-ch

	if data.Error != nil {
		errString := "Error while fetching posts!"
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &errString})
		log.Println(data.Error)
		return
	}

	if len(data.Data.Data.Children) == 0 {
		errString := "Subreddit not found!"
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &errString})
		log.Println(data.Error)
		return
	}

	var redditResp reddit.RedditChildData
	for _, response := range data.Data.Data.Children {
		// Skipping videos, only text and gallery items
		if !response.Data.IsVideo && response.Data.URL != "" &&
			// This is because reddit was using session tokens and only requests with the token could show the Gallery Items
			len(response.Data.Gallery.Items) == 0 {
			redditResp = response
			break
		}
	}

	// TODO: Marhsal json and store all except 0 index value in sqlite3 for subsequent fetching

	link := fmt.Sprintf("[i.redd.it](%s)", redditResp.Data.URL)
	img := &discordgo.MessageEmbedImage{URL: redditResp.Data.URL}

	embed := []*discordgo.MessageEmbed{{
		Author: &discordgo.MessageEmbedAuthor{
			Name: redditResp.Data.SubredditNamePrefixed,
		},
		Title:       redditResp.Data.Title,
		Description: fmt.Sprintf("Posted by u/%s • %s", redditResp.Data.Author, link),
		URL:         fmt.Sprintf("https://www.reddit.com%s", redditResp.Data.PermaLink),
		Color:       0x000000, // Green
		Image:       img,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("⬆️ %d • ⬇️ %d • 💬 %d", redditResp.Data.UpvoteCount, int(float64(redditResp.Data.UpvoteCount)/(redditResp.Data.UpvoteRatio*100)), redditResp.Data.NumComments),
		},
	}}
	session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Embeds: &embed})

}
