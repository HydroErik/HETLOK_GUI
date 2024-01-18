package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"time"

	//"context"
	//"HETLOK_GUI/apiCall"
	"HETLOK_GUI/mongoDrive"
	"log"
	"math/rand"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var u = uint8(rand.Intn(255))
var mongo_uri, userDB, userCol string

var users map[string]mongoDrive.User

var (
	key   = []byte{239, 57, 183, 33, 121, 175, 214, u, 52, 235, 33, 167, 74, 91, 153, 39}
	store = sessions.NewCookieStore(key)
)

var templates = template.Must(template.ParseFiles(
	"../Templates/index.html",
	"../Templates/login.html",
	"../Templates/logout.html",
	"../Templates/pipeDemo.html",
	"../Templates/admin.html",
	"../Templates/users.html",
	"../Templates/userAdd.html",
	"../Templates/userUpdate.html",
	"../Templates/userDelete.html",
))

// Render the provide template string with the passed in data
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	//fmt.Println(data)
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Take in http resopnse writer and set re-validate headers
func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
	return w
}

// Authentication function and re-route
func authenticate(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	if auth, ok := s.Values["authenticated"].(bool); !ok || !auth {
		fmt.Println("Redirecting per auth")
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "hydro-cookie")
		authenticate(w, r, session)
		fn(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	session, _ := store.Get(r, "hydro-cookie")
	var admin bool

	//Check if user had been loaded yet, if not default to not admin
	user, ok := session.Values["user"]
	if !ok {
		admin = false
	} else {
		admin = users[user.(string)].Admin
	}
	renderTemplate(w, "index", map[string]bool{"Admin": admin})
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	session, _ := store.Get(r, "hydro-cookie")
	user, ok := session.Values["user"]
	if ok && users[user.(string)].Admin {
		renderTemplate(w, "admin", users)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	renderTemplate(w, "users", users)
}

func userAddHandler(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Error bool
		Text  string
	}

	w = setHeaders(w)
	switch r.Method {
	case "GET":
		data := Data{
			Error: false,
			Text:  "",
		}
		renderTemplate(w, "userAdd", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			erStr := "Failed to parse from: " + err.Error()
			renderTemplate(w, "userAdd", erStr)
		}
		usrNme := r.PostForm["username"][0]
		pswrdRaw := r.PostForm["password"][0]
		name := r.PostForm["name"][0]
		email := r.PostForm["email"][0]
		_, admin := r.PostForm["admin"]
		newUser := mongoDrive.User{
			Name:     name,
			Username: usrNme,
			Password: pswrdRaw,
			Email:    email,
			Admin:    admin,
		}
		fmt.Println(newUser)

		//Create new client with custome error string on fail to render in browser

		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))
		if err != nil {
			erStr := "Database Connection Failure Error: " + err.Error()
			data := Data{
				Error: true,
				Text:  erStr,
			}
			renderTemplate(w, "userAdd", data)
			return
		}

		//Call to add new user with custome error message to render in browser
		err = mongoDrive.AddUser(userCol, userDB, client, newUser)
		if err != nil {
			erStr := "Failed to update User Db with error: " + err.Error()
			data := Data{
				Error: true,
				Text:  erStr,
			}
			renderTemplate(w, "userAdd", data)
			return
		}

		//update users list to reflect new user
		users, err = mongoDrive.GetUsers(userCol, userDB, client)
		if err != nil {
			erStr := "Failed to update Local user list with error: " + err.Error()
			data := Data{
				Error: true,
				Text:  erStr,
			}
			renderTemplate(w, "userAdd", data)
			return
		}
		//If no errors then asume user added
		data := Data{
			Error: false,
			Text:  "User Added",
		}
		renderTemplate(w, "userAdd", data)

	}

}

func userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	renderTemplate(w, "userUpdate", "")
}

func userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	renderTemplate(w, "userDelete", "")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Login handler Called")
	session, _ := store.Get(r, "hydro-cookie")
	w = setHeaders(w)

	//If already authenticated push to index
	val, ok := session.Values["authenticated"].(bool)
	if ok && val {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	AuthEr, ok := session.Values["authError"]
	if !ok {
		renderTemplate(w, "login", "")
	} else {
		renderTemplate(w, "login", AuthEr)
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "hydro-cookie")
	err := r.ParseForm()
	if err != nil {
		session.Values["authError"] = "Server Error Parsing From Submission:\n" + err.Error()
		session.Save(r, w)
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	usrNme := r.PostForm["username"][0]
	pswrdRaw := r.PostForm["password"][0]
	curUser, ok := users[usrNme]
	if !ok {
		session.Values["authError"] = "Username Not Found"
		session.Save(r, w)
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	pasCrypt := curUser.Password
	err = bcrypt.CompareHashAndPassword([]byte(pasCrypt), []byte(pswrdRaw))
	if err != nil {
		session.Values["authError"] = "Incorect Password"
		session.Save(r, w)
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	session.Values["usrName"] = curUser.Name
	session.Values["authenticated"] = true
	session.Values["user"] = usrNme
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "hydro-cookie")
	auth, ok := session.Values["authenticated"].(bool)
	fmt.Println("Logout Called")
	if auth && ok {
		session.Values["authenticated"] = false
		session.Values["user"] = nil
		session.Values["authError"] = nil
		session.Save(r, w)
		renderTemplate(w, "logout", "")
	} else {
		http.NotFound(w, r)
	}
}

// This loop demonstrates running concurancy and its ease of use
func testLoop() {
	c := 0
	fmt.Printf("Server Running")
	for {

		fmt.Printf(".")
		c++
		time.Sleep(1 * time.Second)
		//Every 4 dots clear stdout
		if c%4 == 0 {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
			fmt.Printf("Server Running")
		}
	}
}

func main() {

	//env file for sensative data
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	mongo_uri = os.Getenv("MONGOSTRING")
	userDB = os.Getenv("USERDB")
	userCol = os.Getenv("USERCOL")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatalf("Couldnt Connect to MongoDB with error:\n%v", err)
	}

	users, err = mongoDrive.GetUsers(userCol, userDB, client)
	if err != nil {
		log.Fatalf("Failed to get user Database with error:\n%v", err)
	}

	//demoPipe, _ = apiCall.TransformerCall(true)

	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/admin/", makeHandler(adminHandler))

	http.HandleFunc("/users", makeHandler(usersHandler))
	http.HandleFunc("/addUser", makeHandler(userAddHandler))
	http.HandleFunc("/updateUser", makeHandler(userUpdateHandler))
	http.HandleFunc("/deleteUser", makeHandler(userDeleteHandler))

	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/validate/", validateHandler)
	http.HandleFunc("/logout/", logoutHandler)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/resources"))))
	http.Handle("/JS/", http.StripPrefix("/JS/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/JS"))))

	//go testLoop() //This is our concurency loop test
	log.Fatal(http.ListenAndServe(":8080", nil))

}
