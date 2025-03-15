package commands

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/reject-Unknown/AnimeQuizBot/bot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Params struct {
	Difficulty bot.Difficulty
	Page       int
}

type Score struct {
	GuildId    string         `bson:"guild_id"`
	UserId     string         `bson:"user_id"`
	Difficulty bot.Difficulty `bson:"difficulty"`
	Score      int            `bson:"score"`
	UserName   string         `bson:"username"`
}

const (
	MAX_PAGE     int = 10
	PAGE_SIZE    int = 10
	DEFAULT_PAGE int = 1
)

func extractParams(options []*discordgo.ApplicationCommandInteractionDataOption) Params {
	var difficulty bot.Difficulty = bot.Difficulty(options[0].IntValue())
	var page int = DEFAULT_PAGE
	if len(options) > 1 {
		page = int(options[1].IntValue())
	}

	return Params{
		Difficulty: difficulty,
		Page:       page,
	}
}

func Leaderboard(botContext *bot.Context) {
	var session *discordgo.Session = botContext.Session
	var interactionCreate *discordgo.InteractionCreate = botContext.InteractionCreate
	var params Params = extractParams(interactionCreate.ApplicationCommandData().Options)

	if params.Page <= 0 {
		session.InteractionRespond(interactionCreate.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Parameter `page` must be positive",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(botContext.GlobalContext.MongoCredentials.ApplyURI))

	if err != nil {
		panic(err.Error())
	}

	collection := client.Database("QuizDB").Collection("Leaderboard")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	options := options.Find()
	options.SetSort(bson.D{{Key: "score", Value: -1}})
	options.SetLimit(int64(MAX_PAGE) * int64(PAGE_SIZE))

	cur, err := collection.Find(ctx, bson.D{
		{Key: "guild_id", Value: botContext.InteractionCreate.GuildID},
		{Key: "difficulty", Value: params.Difficulty},
	}, options)
	if err != nil {
		panic(err.Error())
	}

	var scores []Score
	for cur.Next(ctx) {
		if err := cur.All(ctx, &scores); err != nil {
			log.Fatal(err.Error())
		}
	}

	total_pages := max(1, (PAGE_SIZE+len(scores)-1)/PAGE_SIZE)
	if params.Page > total_pages {
		session.InteractionRespond(interactionCreate.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("`page` parameter value is too large. Max value is %d", total_pages),
			},
		})
	}

	startElement := (params.Page - 1) * 10
	endElement := min((params.Page)*10, len(scores))

	pageScores := scores[startElement:endElement]

	maxLength := 0
	for _, score := range scores {
		maxLength = max(len(score.UserName), maxLength)
	}

	var list string = ""
	for idx, score := range pageScores {
		list += fmt.Sprintf("%d. %s %d\n", startElement+idx+1, score.UserName+strings.Repeat(" ", maxLength-len(score.UserName)), score.Score)
	}

	guild, err := session.Guild(interactionCreate.GuildID)
	if err != nil {
		log.Fatal(err.Error())
	}

	var leaderboard string = ""
	if list != "" {
		leaderboard = fmt.Sprintf("`%s`", list)
	}

	session.InteractionRespond(interactionCreate.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: guild.Name,
					Color: 0xf3a505,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    fmt.Sprintf("AnimeQuiz| Leaderboard [%s]", bot.LevelMap[params.Difficulty]),
						IconURL: guild.IconURL(""),
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Just text from the bottom to hide the insert",
					},
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  fmt.Sprintf("Page %d of %d", params.Page, total_pages),
							Value: leaderboard,
						},
					},
				},
			},
		},
	})

}
