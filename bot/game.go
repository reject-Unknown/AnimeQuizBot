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
	GuildID             string
	CurrentScore        int
	Question            int
}

func NewGame(channelID string, user *discordgo.User, difficulty Difficulty, guildId string) *Game {
	return &Game{
		User:                user,
		CurrentInteraction:  nil,
		PreviousInteraction: nil,
		ChannelID:           channelID,
		Answer:              make(chan string),
		Difficulty:          difficulty,
		GuildID:             guildId,
		CurrentScore:        0,
		Question:            0,
	}
}
