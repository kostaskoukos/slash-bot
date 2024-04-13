package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func iferr(err *error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	godotenv.Load(".env")
	KEY := os.Getenv("BOT_KEY")
	session, err := discordgo.New("Bot " + KEY)
	iferr(&err)
	{
		err := session.Open()
		iferr(&err)
	}
	fmt.Println("hey")
}
