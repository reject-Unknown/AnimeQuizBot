package bot

import (
	"time"

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
	rnd := rand.New(rand.NewSource(uint64(time.Now().Unix())))
	result := []int{}
	found := make(map[int]struct{})
	for range count {
		newValue := rnd.Intn(max) + min
		println(newValue)
		for _, ok := found[newValue]; ok; newValue = rnd.Intn(max) + min {
			_, ok = found[newValue]
		}
		result = append(result, newValue)
		found[newValue] = struct{}{}

	}
	return result
}
