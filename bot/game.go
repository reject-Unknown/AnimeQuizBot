package bot

import (
	"github.com/bwmarrin/discordgo"
)

var LevelMap = map[Difficulty]string{
	EASY:   "Easy",
	MEDIUM: "Medium",
	HARD:   "Hard",
}

type Game struct {
	User                *discordgo.User
	CurrentInteraction  *discordgo.Interaction
	PreviousInteraction *discordgo.Interaction
	Difficulty          Difficulty
	Answer              chan string
	ChannelID           string
	CurrentScore        int
	Question            int
}

func NewGame(channelID string, user *discordgo.User) *Game {
	return &Game{
		User:                user,
		CurrentInteraction:  nil,
		PreviousInteraction: nil,
		ChannelID:           channelID,
		Answer:              make(chan string),
		CurrentScore:        0,
		Question:            0,
	}
}
