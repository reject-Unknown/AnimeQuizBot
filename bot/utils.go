package bot

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/rand"
)

func GetUser(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
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
