package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Pauloo27/searchtube"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func respond(s *discordgo.Session, i *discordgo.Interaction, opts *discordgo.WebhookParams) (*discordgo.Message, error) {
	return s.FollowupMessageCreate(i, true, opts)
}

func fail(s *discordgo.Session, i *discordgo.Interaction, msg string) {
	respond(s, i, &discordgo.WebhookParams{
		Content: msg,
	})
}

func downloadSong(song string, s *discordgo.Session, i *discordgo.Interaction, filename string) *searchtube.SearchResult {
	res, err_s := searchtube.Search(song, 1)
	log.Println("[downloadSong]: found " + song)
	if err_s != nil {
		fail(s, i, "Είμαι πολύ μαύρος για αυτή την εντολή... Ξαναδοκίμασε!")
	}
	if len(res) == 0 {
		fail(s, i, "Δεν μπόρεσα να βρω το τραγούδι... Ξαναδοκίμασε!")
	}
	vid := res[0]
	client := youtube.Client{}
	video, err0 := client.GetVideo(vid.ID)
	if err0 != nil {
		fail(s, i, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
	}

	stream, _, err1 := client.GetStream(video, &video.Formats.Itag(140)[0])
	log.Println("[downloadSong]: found stream")
	if err1 != nil {
		fail(s, i, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
	}
	defer stream.Close()

	file, err2 := os.Create(filename)
	if err2 != nil {
		fail(s, i, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
	}
	defer file.Close()

	_, err3 := io.Copy(file, stream)
	log.Println("[downloadSong]: copying stream...")
	if err3 != nil {
		fail(s, i, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
	}
	return vid
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
	"hey": {
		opts: &discordgo.ApplicationCommand{
			Name:        "hey",
			Description: "Greets the user",
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Γεια σου μαύρε!",
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
						Content: "Είμαι πολύ μαύρος για αυτήν την εντολή... Ξαναδοκίμασε!",
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
						Content: "Είμαι πολύ μαύρος για αυτήν την εντολή... Ξαναδοκίμασε!",
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
	"dl": {
		opts: &discordgo.ApplicationCommand{
			Name:        "dl",
			Description: "Downloads (from YT) and sends an audio file with the song specified",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "song_name",
					Description: "The name of the song you want do download",
					Required:    true,
				},
			},
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			vid := downloadSong(i.ApplicationCommandData().Options[0].StringValue(), s, i.Interaction, "audio.mp3")

			r, err_r := os.Open("audio.mp3")
			if err_r != nil {
				fail(s, i.Interaction, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
			}
			defer r.Close()

			_, err_i := respond(s, i.Interaction, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Your song is %v. Duration: %v. %v", vid.Title, vid.RawDuration, vid.Thumbnail),
				Files: []*discordgo.File{
					{
						Name:        vid.Title + ".mp3",
						ContentType: "audio/mp3",
						Reader:      r,
					},
				},
			})
			if err_i != nil {
				fail(s, i.Interaction, "Δεν μπόρεσα να κατεβάσω το τραγούδι... Ξαναδοκίμασε!")
			}
		},
	},
	"play": {
		opts: &discordgo.ApplicationCommand{
			Name:        "play",
			Description: "Plays the specified song in the voice channel the user is in",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "song_name",
					Description: "The name of the song you want to play",
					Required:    true,
				},
			},
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			disconnect := func() {
				connection, ok := s.VoiceConnections[i.GuildID]
				if ok {
					connection.Disconnect()
				}
			}
			disconnect()

			conn, err := s.ChannelVoiceJoin(i.GuildID, i.ChannelID, false, true)
			if err != nil {
				log.Println(err)
				respond(s, i.Interaction, &discordgo.WebhookParams{
					Content: "Δεν μπόρεσα να μπω στο voice channel... Ξαναδοκίμασε!",
				})
				return
			}
			vid := downloadSong(i.ApplicationCommandData().Options[0].StringValue(), s, i.Interaction, "voice.mp3")

			respond(s, i.Interaction, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Now Playing: %v. Duration: %v. %v", vid.Title, vid.RawDuration, vid.Thumbnail),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Stop",
								Style:    discordgo.DangerButton,
								CustomID: "stop",
							},
						},
					},
				},
			})
			dgvoice.PlayAudioFile(conn, "voice.mp3", make(chan bool))

			respond(s, i.Interaction, &discordgo.WebhookParams{
				Content: "Βγαίνω από το voice channel",
			})

			disconnect()
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
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				if cmd, ok := commands[i.ApplicationCommandData().Name]; ok {
					cmd.handler(s, i)
				}
			case discordgo.InteractionMessageComponent:
				if i.MessageComponentData().CustomID != "stop" {
					return
				}
				connection, ok := s.VoiceConnections[i.GuildID]
				if !ok {
					return
				}

				connection.Disconnect()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Βγαίνω από το voice channel",
					},
				})
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

	os.Remove("audio.mp3")
	os.Remove("voice.mp3")
	log.Println("closing")
}
