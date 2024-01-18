package mongoDrive

import (
	"context"

	//"os"
	//"reflect"

	//"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Name     string
	Email    string
	Password string
	Username string
	Admin    bool
}

// Take in the db and colletion names for user authentication and return a map collection of user objects
// This returned map will be used for main user authentication
func GetUsers(usrColl string, usrDB string, client *mongo.Client) (map[string]User, error) {
	userCol := client.Database(usrDB).Collection(usrColl)
	cur, err := userCol.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var userMap = make(map[string]User)

	for cur.Next(context.TODO()) {
		dbUser := bson.M{}
		err := cur.Decode(&dbUser)
		if err != nil {
			return nil, err
		}
		newUser := User{
			Email:    dbUser["email"].(string),
			Password: dbUser["password"].(string),
			Name:     dbUser["name"].(string),
			Username: dbUser["username"].(string),
			Admin:	  dbUser["admin"].(bool),
		}
		userMap[newUser.Username] = newUser
	}
	return userMap, nil
}
