package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

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
	"meme": {
		opts: &discordgo.ApplicationCommand{
			Name:        "meme",
			Description: "Sends a random meme from reddit",
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			res, err := http.Get("https://meme-api.com/gimme")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "An error occured...",
					},
				})
			}
			data, _ := io.ReadAll(res.Body)

			var meme struct {
				Url string `json:"url"`
			}
			json.Unmarshal(data, &meme)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: meme.Url,
				},
			})
		},
	},
	"joke": {
		opts: &discordgo.ApplicationCommand{
			Name:        "joke",
			Description: "Sends a random joke",
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			res, err := http.Get("https://v2.jokeapi.dev/joke/Any")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "An error occured...",
					},
				})
			}
			data, _ := io.ReadAll(res.Body)

			var joke struct {
				Type     string `json:"type"`
				Joke     string `json:"joke"`
				Setup    string `json:"setup"`
				Delivery string `json:"delivery"`
			}
			json.Unmarshal(data, &joke)

			content := ""
			if joke.Type == "single" {
				content = joke.Joke
			} else {
				content = fmt.Sprintf("%v\n\n%v", joke.Setup, joke.Delivery)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
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
	data := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for name := range commands {
		opts := commands[name].opts
		log.Println(prettyPrint(*opts))
		data = append(data, opts)
	}

	_, err := session.ApplicationCommandBulkOverwrite(APP_KEY, "", data)
	if err != nil {
		log.Fatalf("Couldn't register commands: %v", err)
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
