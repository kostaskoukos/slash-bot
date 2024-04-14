package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func iferr(err error) {
	if err != nil {
		log.Println(err)
	}
}

var (
	TOKEN   string
	APP_KEY string
	session *discordgo.Session
)

func init() {
	log.Println("loading env vars...")
	godotenv.Load(".env")
	TOKEN = os.Getenv("TOKEN")
	APP_KEY = os.Getenv("APP_KEY")

	log.Println("creating bot...")

	var err error
	session, err = discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalf("Cannot create the bot: %v", err)
	}
}

func main() {
	_, err := session.ApplicationCommandBulkOverwrite(APP_KEY, "", []*discordgo.ApplicationCommand{
		{
			Name:        "test",
			Description: "Showcase of a basic slash command",
		},
	})
	if err != nil {
		log.Fatalf("Couldn't register commands: %v", err)
	}
	{
		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		})

		log.Println("starting websocket connection...")
		err := session.Open()
		if err != nil {
			log.Fatalf("Cannot open websocker session: %v", err)
		}
		defer session.Close()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		log.Println("Press Ctrl+C to exit")
		<-stop
	}

	log.Println("closing")
}
