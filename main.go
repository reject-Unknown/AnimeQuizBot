package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/reject-Unknown/AnimeQuizBot/bot"
	"github.com/reject-Unknown/AnimeQuizBot/bot/commands"
)

var (
	Token    string
	User     string
	Password string

	GlobalContext *bot.GlobalContext
)

func Init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&User, "u", "", "Mongo DB Username")
	flag.StringVar(&Password, "p", "", "Mongo DB username password")
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

func GetUser(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		var command string = i.ApplicationCommandData().Name
		if command == "animequiz" {
			context := bot.Context{
				GlobalContext:     GlobalContext,
				Session:           s,
				InteractionCreate: i,
			}

			commands.CharacterQuiz(&context)
		}

	case discordgo.InteractionMessageComponent:
		customId := i.MessageComponentData().CustomID
		switch customId {
		case "1", "2", "3", "4":
			if val, ok := GlobalContext.Games[i.ChannelID]; ok {
				user := GetUser(i.Interaction)
				if val.User.ID == user.ID {
					val.PreviousInteraction = val.CurrentInteraction
					val.CurrentInteraction = i.Interaction
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

	GlobalContext = &bot.GlobalContext{
		Games: make(map[string]*bot.Game),
		Data:  bot.LoadData(User, Password),
	}

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
