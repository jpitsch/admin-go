package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/http"
)

type User struct {
	Name     string `bson:"name"`
	Password string `bson:"password"`
}

type UserRole struct {
	Rolename string `bson:"rolename"`
	Approle  string `bson:"approle"`
}

var testuri = "mongodb://pitsch:test@ds029051.mongolab.com:29051/pitsch_test"

var templates = template.Must(template.ParseFiles("login.html", "view.html"))

//Setting up the database variables
var (
	mgoSession   *mgo.Session
	databaseName = "pitsch_test"
)

//Only get the copy of the session
func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(testuri)
		if err != nil {
			fmt.Println("Error connecting to Mongo: ", err)
		}
	}
	return mgoSession.Copy()
}

//var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
func loadUser(user string) User {
	session := getSession()
	defer session.Close()

	users := session.DB("pitsch_test").C("users")

	userresult := User{}
	err := users.Find(bson.M{"name": user}).One(&userresult)
	if err != nil {
		fmt.Println("No such user: ")
	}

	return userresult
}

//Example of passing reference of the user type
func saveUser(u User) {
	session := getSession()
	defer session.Close()

	c := session.DB("pitsch_test").C("users")
	err := c.Insert(&User{u.Name, u.Password})

	if err != nil {
		fmt.Println("Error while trying to save to Mongo.")
	}
}

func createUserApi(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	u := User{}

	err := decoder.Decode(&u)
	if err != nil {
		fmt.Println("Error trying to decode JSON request for create user.")
	}

	saveUser(u)
}

func getUserApi(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	u := User{}

	err := decoder.Decode(&u)
	if err != nil {
		fmt.Println("Error trying to decode JSON request for get user.")
	}

	u = loadUser(u.Name)
	res, err := json.Marshal(u)

	if err != nil {
		fmt.Println("Error trying to create JSON response")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func saveUserHandler(w http.ResponseWriter, r *http.Request) {
	uname := r.FormValue("user")
	upass := r.FormValue("password")

	u := User{Name: uname, Password: upass}
	saveUser(u)

	http.Redirect(w, r, "/view/"+uname, http.StatusFound)
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	u := User{}

	t, _ := template.ParseFiles("create.html")
	t.Execute(w, u)
}

func viewUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Path[len("/view/"):]
	u := loadUser(username)

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, u)
}

func main() {
	http.HandleFunc("/rest/create/", createUserApi)
	http.HandleFunc("/rest/getuser/", getUserApi)
	http.HandleFunc("/view/", viewUserHandler)
	http.HandleFunc("/edit/", editUserHandler)
	http.HandleFunc("/save/", saveUserHandler)
	http.ListenAndServe(":8080", nil)
}
