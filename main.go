package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	TOKEN   string
	APP_KEY string
	session *discordgo.Session
)

type Command struct {
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
	opts    *discordgo.ApplicationCommand
}

var commands = map[string]Command{
	"test": {
		opts: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test a simple bot command",
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		},
	},
}

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
	for name := range commands {
		opts := commands[name].opts
		_, err := session.ApplicationCommandBulkOverwrite(APP_KEY, "", []*discordgo.ApplicationCommand{opts})
		if err != nil {
			log.Fatalf("couldnt register command '%v': %v", name, err)
		}
	}
	{
		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if cmd, ok := commands[i.ApplicationCommandData().Name]; ok {
				cmd.handler(s, i)
			}
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
