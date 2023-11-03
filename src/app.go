package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	//"context"
	"HETLOK_GUI/APICall"
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

var users map[string]mongoDrive.User

var demoPipe APICall.PipeObj

var (
	key   = []byte{239, 57, 183, 33, 121, 175, 214, u, 52, 235, 33, 167, 74, 91, 153, 39}
	store = sessions.NewCookieStore(key)
)

var templates = template.Must(template.ParseFiles(
	"../Templates/index.html",
	"../Templates/login.html",
	"../Templates/logout.html",
	"../Templates/pipeDemo.html",
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

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "hydro-cookie")

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Redirecting per auth")
			http.Redirect(w, r, "/login/", http.StatusFound)
		}
		fn(w, r)
	}
}

func makeDemoHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "hydro-cookie")

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Redirecting per auth")
			http.Redirect(w, r, "/login/", http.StatusFound)
		}
		fn(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("index handler called")
	//session, _ := store.Get(r, "hydro-cookie")
	w = setHeaders(w)
	renderTemplate(w, "index", "")
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
	//TODO fill in validation logic
	//Set auth to true
	session.Values["authenticated"] = true
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "hydro-cookie")
	auth, ok := session.Values["authenticated"].(bool)
	fmt.Println("Logout Called")
	if auth && ok {
		session.Values["authenticated"] = false
		session.Save(r, w)
		renderTemplate(w, "logout", "")
	} else {
		http.NotFound(w, r)
	}
}

// Handle a demo selection screen
func demoSelectionHandler(w http.ResponseWriter, r *http.Request) {
	
	renderTemplate(w, "Pipe Selector", data)
}

func demoHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "pipeDemo", data)
}

//This loop demonstrates running concurancy and its ease of use
/*
func testLoop() {
	for {
		fmt.Println("Test")
		time.Sleep(1500000000)
	}
}
*/

func main() {

	//env file for sensative data
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	mongo_uri := os.Getenv("MONGOSTRING")
	userDB := os.Getenv("USERDB")
	userCol := os.Getenv("USERCOL")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatalf("Couldnt Connect to MongoDB with error:\n%v", err)
	}

	users, err = mongoDrive.GetUsers(userCol, userDB, client)
	if err != nil {
		log.Fatalf("Failed to get user Database with error:\n%v", err)
	}

	demoPipe = APICall.TransformerCall()

	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/pipes/", makeDemoHandler(demoHandler))
	http.HandleFunc("/select/", makeDemoHandler(demoSelectionHandler))
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/validate/", validateHandler)
	http.HandleFunc("/logout/", logoutHandler)

	//go testLoop() This is our concurency loop test
	log.Fatal(http.ListenAndServe(":8080", nil))

}
