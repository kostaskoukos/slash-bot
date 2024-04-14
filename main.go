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

func main() {
	log.Println("loading env vars...")
	godotenv.Load(".env")
	BOT_KEY := os.Getenv("BOT_KEY")
	APP_KEY := os.Getenv("APP_KEY")
	log.Println("creating bot...")
	session, err := discordgo.New("Bot " + BOT_KEY)
	iferr(err)
	{
		cmds, err := session.ApplicationCommandBulkOverwrite(APP_KEY, "", []*discordgo.ApplicationCommand{
			{
				Name:        "test",
				Description: "Showcase of a basic slash command",
			},
		})

		session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Println("interaction with: ", i.Member)
		})
		log.Println("cmds: ", cmds)
		iferr(err)
	}
	{
		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		})

		err := session.Open()
		iferr(err)
		log.Println("starting websocket connection...")

		defer session.Close()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		log.Println("Press Ctrl+C to exit")
		<-stop
	}

	log.Println("closing")
}
