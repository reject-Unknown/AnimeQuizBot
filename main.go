package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/reject-Unknown/AnimeQuizBot/bot"
)

var (
	Token    string
	User     string
	Password string
	Games    map[string]*bot.Game = make(map[string]*bot.Game)
	Data     bot.CharactersData
)

func Init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&User, "u", "", "Mongo DB Username")
	flag.StringVar(&Password, "p", "", "Mongo DB Password UserName")
	flag.Parse()
}

func mapsToCommandChoices() []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice
	for mapValue, mapName := range bot.LevelMap {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  mapName,
			Value: mapValue,
		})
	}
	return choices
}

func GetUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func changeMessage(message *discordgo.MessageSend, colour int) {
	emded := message.Embeds[0]
	emded.Color = colour
	message.Components = []discordgo.MessageComponent{}
}

func animeQuiz(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	difficulty := bot.Level(options[0].Value.(float64))
	user := GetUser(i)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Welcome %s to %s Anime Quiz", user.Mention(), bot.LevelMap[difficulty]),
		},
	})

	game := bot.NewGame(i.ChannelID, user.ID)
	Games[game.ChannelID] = game
	levelData := Data[difficulty]
loop:
	for {
		game.Question++
		rndixs := RandUniqueNumbers(0, len(levelData), 4)
		correctAnswer := rand.Intn(len(rndixs)) + 1
		charachers := []*bot.Character{}
		for _, val := range rndixs {
			charachers = append(charachers, levelData[val])
		}

		var pickList string = fmt.Sprintf("**[1]** %s\n**[2]** %s\n**[3]** %s\n**[4]** %s\n", charachers[0].Name, charachers[1].Name, charachers[2].Name, charachers[3].Name)
		message := discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: fmt.Sprintf("\U00002753 Question #%d \n Do you know who is it?", game.Question),
					Image: &discordgo.MessageEmbedImage{
						URL: charachers[correctAnswer-1].ImageUrl,
					},
					Color: 2458803,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    fmt.Sprintf("%s | AnimeQuiz [%s]", user.GlobalName, bot.LevelMap[difficulty]),
						IconURL: user.AvatarURL(""),
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Просто текст внизу чтобы расширить Embed",
					},
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Choose one of the four",
							Value: pickList,
						},
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "1️⃣",
							},
							CustomID: "1",
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "2️⃣",
							},
							CustomID: "2",
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "3️⃣",
							},
							CustomID: "3",
						},
						discordgo.Button{
							Style: discordgo.SecondaryButton,
							Emoji: &discordgo.ComponentEmoji{
								Name: "4️⃣",
							},
							CustomID: "4",
						},
					},
				},
			},
		}
		sendedMessage, _ := s.ChannelMessageSendComplex(i.ChannelID, &message)

		select {
		case res := <-game.Answer:
			if res == strconv.Itoa(correctAnswer) {
				game.CurrentScore += 100
				changeMessage(&message, 53760)
				s.ChannelMessageEditComplex(
					&discordgo.MessageEdit{
						Embeds:     &message.Embeds,
						Components: &message.Components,
						ID:         sendedMessage.ID,
						Channel:    sendedMessage.ChannelID,
					},
				)
				s.InteractionRespond(
					game.Interaction.Interaction,
					&discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("✅ Correct! Your Score %d", game.CurrentScore),
						},
					},
				)
			} else {
				changeMessage(&message, 13238272)
				s.ChannelMessageEditComplex(
					&discordgo.MessageEdit{
						Embeds:     &message.Embeds,
						Components: &message.Components,
						ID:         sendedMessage.ID,
						Channel:    sendedMessage.ChannelID,
					},
				)
				s.InteractionRespond(
					game.Interaction.Interaction,
					&discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("Ты проебал! Your Score %d", game.CurrentScore),
						},
					},
				)
				break loop
			}
		case <-time.After(10 * time.Second):
			s.ChannelMessageSend(i.ChannelID, "Слишком долго думаешь, пиздуй отсюда")
			break loop
		}

	}

	delete(Games, game.ChannelID)
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		var command string = i.ApplicationCommandData().Name
		if command == "animequiz" {
			animeQuiz(s, i)
		}
	case discordgo.InteractionMessageComponent:
		customId := i.MessageComponentData().CustomID
		switch customId {
		case "1", "2", "3", "4":
			if val, ok := Games[i.ChannelID]; ok {
				user := GetUser(i)
				if val.UserID == user.ID {
					val.Interaction = i
					val.Answer <- customId
				}
			}
		}
	}

}

func RandUniqueNumbers(min int, max int, count int) []int {
	result := []int{}
	found := make(map[int]struct{})
	for range count {
		newValue := rand.Intn(max) + min
		for _, ok := found[newValue]; ok; newValue = rand.Intn(max) + min {
		}
		result = append(result, newValue)
		found[newValue] = struct{}{}

	}
	return result
}

func main() {
	Init()
	Data = bot.LoadData(User, Password)
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(interactionCreate)

	_, err = dg.ApplicationCommandCreate("728036143785967706", "", &discordgo.ApplicationCommand{
		Name:        "animequiz",
		Description: "Start Anime Quiz game to guess characters for pictures",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "difficult",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Description: "Map to display",
				Required:    true,
				Choices:     mapsToCommandChoices(),
			},
		},
	})
	if err != nil {
		println(err.Error())
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}
