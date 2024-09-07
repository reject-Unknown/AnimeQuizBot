package bot

import "github.com/bwmarrin/discordgo"

var LevelMap = map[Level]string{
	EASY:   "Easy",
	MEDIUM: "Medium",
	HARD:   "Hard",
}

type Game struct {
	ChannelID    string
	UserID       string
	Answer       chan string
	Interaction  *discordgo.InteractionCreate
	CurrentScore int
	Question     int
}

func NewGame(channelID string, userID string) *Game {
	return &Game{
		ChannelID:    channelID,
		UserID:       userID,
		Answer:       make(chan string),
		Interaction:  nil,
		CurrentScore: 0,
		Question:     0,
	}
}
