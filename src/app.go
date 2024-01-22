package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	//"context"
	//"HETLOK_GUI/apiCall"
	//"HETLOK_GUI/mongoDrive"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var u = uint8(rand.Intn(255))
var (
	key   = []byte{239, 57, 183, 33, 121, 175, 214, u, 52, 235, 33, 167, 74, 91, 153, 39}
	store = sessions.NewCookieStore(key)
)

var templates = template.Must(template.ParseFiles(
	"../Templates/index.html",
	"../Templates/login.html",
	"../Templates/logout.html",
))

var conf = &oauth2.Config{
	ClientID:     "548937884118-al1bjls2dck7600t2dl3p9ehgbg5atl8.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:8080/validate/",
	Scopes: []string{
		"openid",
		"email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

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
	renderTemplate(w, "index", session.Values["name"])
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Login handler Called")
	session, _ := store.Get(r, "hydro-cookie")
	w = setHeaders(w)

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state")

	//If already authenticated push to index
	val, ok := session.Values["authenticated"].(bool)
	if ok && val {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	renderTemplate(w, "login", url)
}

// ?state=state&code=4%2F0AfJohXmWhexYzACmhbR3vtNaWQTuUnmyMDT0K9jQBmoHH23rNjyYkLiqiWPE_f7ApJw_YQ&scope=openid&authuser=0&prompt=consent
func oauthValidate(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "hydro-cookie")
	code := r.URL.Query().Get("code")
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info using token
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo?alt=json", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode user info JSON
	var userinfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userinfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["email"] = userinfo["email"]
	session.Values["name"] = userinfo["given_name"]
	session.Values["authenticated"] = true
	session.Values["accessToken"] = token.AccessToken
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "hydro-cookie")
	auth, ok := session.Values["authenticated"].(bool)
	fmt.Println("Logout Called")
	if auth && ok {
		//Google logout URL
		url := "https://accounts.google.com/o/oauth2/revoke?token=" + session.Values["accessToken"].(string)

		//logout of google
		resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
		if err != nil {
			http.NotFound(w, r)
		}
		defer resp.Body.Close()

		//app control log out
		session.Values["authenticated"] = false
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
	//demoPipe, _ = apiCall.TransformerCall(true)

	http.HandleFunc("/", makeHandler(indexHandler))

	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/validate/", oauthValidate)
	http.HandleFunc("/logout/", logoutHandler)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/resources"))))
	http.Handle("/JS/", http.StripPrefix("/JS/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/JS"))))

	//go testLoop() //This is our concurency loop test
	log.Fatal(http.ListenAndServe(":8080", nil))

}
