package bot

import "github.com/bwmarrin/discordgo"

type GlobalContext struct {
	Games map[string]*Game
	Data  map[Difficulty][]*Character
}

type Context struct {
	GlobalContext     *GlobalContext
	Session           *discordgo.Session
	InteractionCreate *discordgo.InteractionCreate
}
