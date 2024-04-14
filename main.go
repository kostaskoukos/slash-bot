package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func iferr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	godotenv.Load(".env")
	BOT_KEY := os.Getenv("BOT_KEY")
	APP_KEY := os.Getenv("APP_KEY")
	session, err := discordgo.New("Bot " + BOT_KEY)
	iferr(err)
	{
		cmds, err := session.ApplicationCommandBulkOverwrite(APP_KEY, "", []*discordgo.ApplicationCommand{
			{
				Name:        "test",
				Description: "Showcase of a basic slash command",
			},
		})
		fmt.Println("cmds:", cmds)
		iferr(err)
	}
	{
		err := session.Open()
		iferr(err)
	}
	fmt.Println("hey")
}
