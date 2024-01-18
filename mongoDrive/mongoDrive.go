package mongoDrive

import (
	"context"

	//"os"
	//"reflect"

	//"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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
			Admin:    dbUser["admin"].(bool),
		}
		userMap[newUser.Username] = newUser
	}
	return userMap, nil
}

// Password encrypter takes raw string and returns new password
func EncryptPass(old_pass string) (string, error) {
	//encoded := base64.StdEncoding.EncodeToString([]byte(old_pass))
	encoded, err := bcrypt.GenerateFromPassword([]byte(old_pass), bcrypt.DefaultCost)
	return string(encoded), err
}


// Take in DB connection parameters and new user User struct issue inject 1 to user DB
// Return DB error if failure
func AddUser(usrColl string, usrDB string, client *mongo.Client, newUser User) error {
	userCol := client.Database(usrDB).Collection(usrColl)
	encrypted, err := EncryptPass(newUser.Password)
	if err != nil {
		return err
	}
	newUser.Password = encrypted

	_, err = userCol.InsertOne(context.TODO(), newUser)
	if err != nil {
		return err
	}

	return nil
}
