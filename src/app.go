package main

import (
	"apiCall"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
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
var hconf oauth2.Config
var conf = &hconf
var Clients []interface{}

var templates = template.Must(template.ParseFiles(
	"../Templates/index.html",
	"../Templates/login.html",
	"../Templates/logout.html",
	"../Templates/clients.html",
	"../Templates/addClient.html",
	"../Templates/editClient.html",
))

var TimeZone map[float64]string

type ClData struct {
	Er     bool
	ErM    string
	MapInt []interface{}
	TZ     map[float64]string
}

// Function handles calling the DB for the client list
// Sorts by clientID sets global and returns nil or error
func setClients() error {
	var err error
	Clients, err = apiCall.QueryDB("client", "a", "", "")
	if err != nil {
		return err
	}

	sort.Slice(Clients, func(i, j int) bool {
		return int(Clients[i].(map[string]interface{})["clientId"].(float64)) < int(Clients[j].(map[string]interface{})["clientId"].(float64))
	})
	return nil
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

// Add client handler
func addClientHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	type aData struct {
		Er  bool
		ErM string
		TZ  map[float64]string
	}
	data := aData{Er: false, ErM: "", TZ: TimeZone}
	switch r.Method {
	case "GET":
		//return the add client form
		renderTemplate(w, "addClient", data)
		return
	case "POST":
		r.ParseForm()

		longName := r.PostForm["new-long-name"][0]
		shortName := r.PostForm["new-short-name"][0]
		tzd, _ := strconv.Atoi(r.PostForm["new-timezone-id"][0])
		timezondId := float64(tzd)
		notes := r.PostForm["new-notes"][0]
		_, ok := r.PostForm["new-enabled"]
		var isEnabled bool
		if ok {
			isEnabled = true
		} else {
			isEnabled = false
		}

		fmt.Printf("New Client name: %s\nshort name: %s\ntimezoneId: %f\nnotes: %s\nenabled: %t\n", longName, shortName, timezondId, notes, isEnabled)
		//TODO
		//We need to make the add client API call with the given data
		cData := ClData{
			Er:     false,
			ErM:    "",
			MapInt: Clients,
			TZ:     TimeZone,
		}
		err := setClients()
		if err != nil {
			data.Er = true
			data.ErM = err.Error()
			renderTemplate(w, "clients", cData)
			return
		}
		cData.MapInt = Clients
		renderTemplate(w, "clients", cData)
		return
	default:
		//Some more fun error handling for that wtf case
		cData := ClData{
			Er:     true,
			ErM:    "How the fuck did you get here!?!",
			MapInt: Clients,
			TZ:     TimeZone,
		}
		renderTemplate(w, "clients", cData)
		return
	}
}

// Handle delete client request
func deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	i := r.URL.Query()["index"][0]
	n, _ := strconv.Atoi(i)
	data := ClData{
		Er:     false,
		ErM:    "",
		MapInt: []interface{}{},
		TZ:     TimeZone,
	}

	fmt.Printf("CLient to delete:\n%s", Clients[n])
	//TODO: Issue delete client API call
	//Await API constuction

	err := setClients()
	if err != nil {
		data.Er = true
		data.ErM = err.Error()
		renderTemplate(w, "clients", data)
		return
	}
	data.MapInt = Clients

	renderTemplate(w, "clients", data)
}

// Handle editing of current client
func editClientHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	i := r.URL.Query()["index"][0]
	n, _ := strconv.Atoi(i)
	curTZID := Clients[n].(map[string]interface{})["timezoneId"].(float64)
	data := struct {
		Er     bool
		ErM    string
		Ind    int
		MapInt interface{}
		TZ     map[float64]string
		CurInt float64
	}{
		Er:     false,
		ErM:    "",
		Ind:    n,
		MapInt: Clients[n],
		TZ:     TimeZone,
		CurInt: curTZID,
	}

	switch r.Method {
	case "GET":
		renderTemplate(w, "editClient", data)
		return
	case "POST":
		r.ParseForm()
		cli := Clients[n].(map[string]interface{})
		cli["name"] = r.PostForm["long-name"][0]
		cli["shortName"] = r.PostForm["short-name"][0]
		tzd, _ := strconv.Atoi(r.PostForm["timezone-id"][0])
		cli["timezondId"] = float64(tzd)
		cli["notes"] = r.PostForm["notes"][0]
		_, ok := r.PostForm["enabled"]
		if ok {
			cli["isEnabled"] = true
		} else {
			cli["isEnabled"] = false
		}
		fmt.Printf("Edited client:\n%s", cli)
		//TODO make edit client API call
		//
		setClients()
		cData := ClData{
			Er:     false,
			ErM:    "",
			MapInt: Clients,
			TZ:     TimeZone,
		}
		renderTemplate(w, "clients", cData)
	default:
		//Return error dont know how we got here but fuck lets handle it
		cData := ClData{
			Er:     true,
			ErM:    "How the fuck did you get here!?!",
			MapInt: Clients,
			TZ:     TimeZone,
		}
		renderTemplate(w, "clients", cData)
	}
}

func clientsHandler(w http.ResponseWriter, r *http.Request) {
	w = setHeaders(w)
	var err error
	data := ClData{
		Er:     false,
		ErM:    "",
		MapInt: []interface{}{},
		TZ:     TimeZone,
	}
	err = setClients()
	if err != nil {
		data.Er = true
		data.ErM = err.Error()
		renderTemplate(w, "clients", data)
		return
	}
	data.MapInt = Clients
	renderTemplate(w, "clients", data)
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
	hconf = oauth2.Config{
		ClientID:     os.Getenv("GID"),
		ClientSecret: os.Getenv("GSEC"),
		RedirectURL:  os.Getenv("RDRCT"),
		Scopes: []string{
			"openid",
			"email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	//TODO this can be an API call to stay consistent
	TimeZone = map[float64]string{
		1: "US_EASTERN",
		2: "US_CENTRAL",
		3: "US_ARIZONA",
		4: "US_MOUNTAIN",
		5: "US_PACIFIC",
		6: "US_ALASKA",
		7: "US_HAWAII",
	}
	//demoPipe, _ = apiCall.TransformerCall(true)

	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/clients", makeHandler(clientsHandler))
	http.HandleFunc("/addClient", makeHandler(addClientHandler))
	http.HandleFunc("/deleteClient", makeHandler(deleteClientHandler))
	http.HandleFunc("/editClient", makeHandler(editClientHandler))

	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/validate/", oauthValidate)
	http.HandleFunc("/logout/", logoutHandler)

	//Serve out local files for resources
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/resources"))))
	http.Handle("/JS/", http.StripPrefix("/JS/", http.FileServer(http.Dir("C:/Users/esunb/Documents/github/HETLOK_GUI/JS"))))

	//go testLoop() //This is our concurency loop test
	log.Fatal(http.ListenAndServe(":8888", nil))

}
