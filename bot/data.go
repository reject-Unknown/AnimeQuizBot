package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Character struct {
	Name     string `bson:"name"`
	ImageUrl string `bson:"image_url"`
	AnimeUrl string `bson:"anime_url"`
	Id       int    `bson:"id"`
}

var (
	CONNECTION_STR string = "mongodb://%s:%s@127.0.0.1:22222/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.3.0"
)

type Difficulty int

const (
	EASY Difficulty = iota
	MEDIUM
	HARD
)

type CharactersData map[Difficulty][]*Character

func LoadData(user string, password string) CharactersData {
	data := make(CharactersData)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf(CONNECTION_STR, user, password)))

	if err != nil {
		panic(err.Error())
	}

	for idx := range LevelMap {
		fmt.Fprintf(os.Stdout, "Index: %d\n", idx)
		collection := client.Database("QuizDB").Collection(fmt.Sprintf("CharactersLevel%d", idx+1))
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cur, err := collection.Find(ctx, bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		var level_characters []*Character = []*Character{}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var result Character
			if err := cur.Decode(&result); err != nil {
				log.Fatal(err)
			}
			level_characters = append(level_characters, &result)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		data[idx] = level_characters
	}
	return data
}
