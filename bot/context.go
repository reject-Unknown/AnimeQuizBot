package bot

type GlobalContext struct {
	Games map[string]*Game
	Data  map[Level][]*Character
}

type Context struct {
	GlobalContext *GlobalContext
}
