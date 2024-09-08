package commands

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/reject-Unknown/AnimeQuizBot/bot"
	"golang.org/x/exp/rand"
)

const (
	QUESTION_SCORE int = 100
)

func getEditMessageOnCorrectAnswer(message *discordgo.Message) *discordgo.MessageEdit {
	emded := message.Embeds[0]
	emded.Color = 53760
	message.Components = []discordgo.MessageComponent{}
	return &discordgo.MessageEdit{
		Embeds:     &message.Embeds,
		Components: &message.Components,
		ID:         message.ID,
		Channel:    message.ChannelID,
	}
}

func getEditMessageOnGameOver(message *discordgo.Message) *discordgo.MessageEdit {
	emded := message.Embeds[0]
	emded.Color = 13238272
	message.Components = []discordgo.MessageComponent{}
	return &discordgo.MessageEdit{
		Embeds:     &message.Embeds,
		Components: &message.Components,
		ID:         message.ID,
		Channel:    message.ChannelID,
	}
}

func timeoutGameOverMessage(game *bot.Game) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: ":alarm_clock: Time is over!",
				Color: 0xf3a505,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    fmt.Sprintf("%s | AnimeQuiz [%s]", game.User.GlobalName, bot.LevelMap[game.Difficulty]),
					IconURL: game.User.AvatarURL(""),
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Just text from the bottom to hide the insert",
				},

				Description: fmt.Sprintf("**Your total score: `%d`**", game.CurrentScore),
			},
		},
	}
}

func failGameOverMessage(game *bot.Game) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: ":x: Game Over! Incorrect answer",
				Color: 0xf3a505,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    fmt.Sprintf("%s | AnimeQuiz [%s]", game.User.GlobalName, bot.LevelMap[game.Difficulty]),
					IconURL: game.User.AvatarURL(""),
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Just text from the bottom to hide the insert",
				},

				Description: fmt.Sprintf("**Your total score: `%d`**", game.CurrentScore),
			},
		},
	}
}

func makeGuessMessage(characters []*bot.Character, game *bot.Game, correctNumber int) *discordgo.MessageSend {
	var pickList string = fmt.Sprintf("**[1]** %s\n**[2]** %s\n**[3]** %s\n**[4]** %s\n", characters[0].Name, characters[1].Name, characters[2].Name, characters[3].Name)
	var user *discordgo.User = game.User
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("\U00002753 Question #%d \n Do you know who is it?", game.Question),
				Image: &discordgo.MessageEmbedImage{
					URL: characters[correctNumber-1].ImageUrl,
				},
				Color: 2458803,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    fmt.Sprintf("%s | AnimeQuiz [%s]", user.GlobalName, bot.LevelMap[game.Difficulty]),
					IconURL: user.AvatarURL(""),
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Just text from the bottom to hide the insert",
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
}

func CharacterQuiz(context *bot.Context) {
	var session *discordgo.Session = context.Session
	var interactionCreate *discordgo.InteractionCreate = context.InteractionCreate

	var difficulty bot.Difficulty = bot.Difficulty(interactionCreate.ApplicationCommandData().Options[0].IntValue())
	var user *discordgo.User = bot.GetUser(interactionCreate.Interaction)

	err := session.InteractionRespond(interactionCreate.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Welcome %s to %s Anime Quiz", user.Mention(), bot.LevelMap[difficulty]),
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	var game *bot.Game = bot.NewGame(interactionCreate.ChannelID, user)
	context.GlobalContext.Games[game.ChannelID] = game

	defer delete(context.GlobalContext.Games, game.ChannelID)
	var gameData []*bot.Character = context.GlobalContext.Data[difficulty]

	var gameOver bool = false
	for !gameOver {
		game.Question++
		rndixs := bot.RandUniqueNumbers(0, len(gameData), 4)
		characters := []*bot.Character{}
		for _, val := range rndixs {
			characters = append(characters, gameData[val])
		}

		correctAnswer := rand.Intn(len(rndixs)) + 1
		var message *discordgo.MessageSend = makeGuessMessage(characters, game, correctAnswer)
		sendedMessage, err := session.ChannelMessageSendComplex(interactionCreate.ChannelID, message)
		if err != nil {
			log.Fatal(err)
		}

		select {
		case userAnswer := <-game.Answer:
			if game.PreviousInteraction != nil {
				session.InteractionResponseDelete(game.PreviousInteraction)
			}

			if userAnswer == strconv.Itoa(correctAnswer) {
				game.CurrentScore += QUESTION_SCORE
				_, err = session.ChannelMessageEditComplex(getEditMessageOnCorrectAnswer(sendedMessage))
				if err != nil {
					log.Fatal(err.Error())
				}

				err = session.InteractionRespond(
					game.CurrentInteraction,
					&discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("✅ Correct! Your Score %d", game.CurrentScore),
						},
					},
				)

				if err != nil {
					log.Fatal(err.Error())
				}

			} else {
				_, err = session.ChannelMessageEditComplex(getEditMessageOnGameOver(sendedMessage))
				if err != nil {
					log.Fatal(err.Error())
				}

				err = session.InteractionRespond(
					game.CurrentInteraction,
					&discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseDeferredMessageUpdate,
					},
				)

				if err != nil {
					log.Fatal(err.Error())
				}

				session.ChannelMessageSendComplex(interactionCreate.ChannelID, failGameOverMessage(game))
				gameOver = true
			}

		case <-time.After(10 * time.Second):
			if game.CurrentInteraction != nil {
				session.InteractionResponseDelete(game.CurrentInteraction)
			}
			_, err = session.ChannelMessageEditComplex(getEditMessageOnGameOver(sendedMessage))
			if err != nil {
				log.Fatal(err.Error())
			}
			session.ChannelMessageSendComplex(interactionCreate.ChannelID, timeoutGameOverMessage(game))
			gameOver = true
		}
	}
}
