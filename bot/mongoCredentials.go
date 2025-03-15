package bot

import "fmt"

const (
	APPLY_URI string = "mongodb://%s:%s@127.0.0.1:22222/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.3.0"
)

type MongoCredentials struct {
	ApplyURI string
}

func NewMongoCredentials(user string, password string) *MongoCredentials {
	return &MongoCredentials{
		ApplyURI: fmt.Sprintf(APPLY_URI, user, password),
	}
}
