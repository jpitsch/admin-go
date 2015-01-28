package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/http"
	"strconv"
	//"time"
)

type User struct {
	Firstname  string `bson:"firstname"`
	Lastname   string `bson:"lastname"`
	Password   string `bson:"password"`
	Email      string `bson:"email"`
	Login      string `bson:"login"`
	IsAdmin    bool   `bson:"isadmin"`
	IsDisabled bool   `bson:"isdisabled"`
}

type UserSession struct {
	SessionKey string `bson:"sessionkey"`
	UserId     string `bson:"userid"`
}

type Category struct {
	Name string
}

type UserRole struct {
	Rolename string `bson:"rolename"`
	Approle  string `bson:"approle"`
}

type Option struct {
	Value, Id, Text string
	Selected        bool
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

func withCollection(collection string, s func(*mgo.Collection) error) error {
	session := getSession()
	defer session.Close()
	c := session.DB(databaseName).C(collection)
	return s(c)
}

//var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
func loadUser(user string, criteria string) User {
	session := getSession()
	defer session.Close()

	users := session.DB("pitsch_test").C("users")

	userresult := User{}
	err := users.Find(bson.M{criteria: user}).One(&userresult)
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
	//err := c.Insert(&User{u.Login, u.Password})
	err := c.Insert(u)

	if err != nil {
		fmt.Println("Error while trying to save to Mongo.")
	}
}

func updateUser(u User) {
	session := getSession()
	defer session.Close()

	//users := session.DB("pitsch_test").C("users")

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

	u = loadUser(u.Login, "login")
	res, err := json.Marshal(u)

	if err != nil {
		fmt.Println("Error trying to create JSON response")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func SearchUser(q interface{}, skip int, limit int) (searchResults []User, searchErr string) {
	searchErr = ""
	searchResults = []User{}
	query := func(c *mgo.Collection) error {
		fn := c.Find(q).Skip(skip).Limit(limit).All(&searchResults)
		if limit < 0 {
			fn = c.Find(q).Skip(skip).All(&searchResults)
		}
		return fn
	}
	search := func() error {
		return withCollection("user", query)
	}
	err := search()
	if err != nil {
		searchErr = "Database Error"
	}
	return
}

func GetAllUsers() {

}

func GetUserByLogin(login string, skip int, limit int) (searchResults []User, searchErr string) {
	searchResults, searchErr = SearchUser(bson.M{"login": bson.RegEx{"^" + login, "i"}}, skip, limit)
	return
}

func saveUserHandler(w http.ResponseWriter, r *http.Request) {
	isadmin, err := strconv.ParseBool(r.FormValue("isadmin"))
	if err != nil {
		fmt.Println("Error while trying to parse bool.")
	}

	u := User{
		Login:      r.FormValue("login"),
		Password:   r.FormValue("password"),
		Firstname:  r.FormValue("firstname"),
		Lastname:   r.FormValue("lastname"),
		Email:      r.FormValue("email"),
		IsAdmin:    isadmin,
		IsDisabled: false}
	saveUser(u)

	http.Redirect(w, r, "/view/"+r.FormValue("login"), http.StatusFound)
}

func editUserHandler(w http.ResponseWriter, r *http.Request) {
	u := User{}

	t, _ := template.ParseFiles("create.html")
	t.Execute(w, u)
}

func viewUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Path[len("/view/"):]
	u := loadUser(username, "login")

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, u)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	pass := r.FormValue("password")

	u := loadUser(login, "Login")
	if u.Password != pass {
		http.Redirect(w, r, "/login/", http.StatusForbidden)
	}
}

func displayLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	u := User{}

	t, _ := template.ParseFiles("login.html")
	t.Execute(w, u)
}

func displayCreateUserPageHandler(w http.ResponseWriter, r *http.Request) {
	u := User{}

	t, _ := template.ParseFiles("create.html")
	t.Execute(w, u)
}

func main() {
	http.HandleFunc("/rest/create/", createUserApi)
	http.HandleFunc("/rest/getuser/", getUserApi)
	http.HandleFunc("/home/", displayLoginPageHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/view/", viewUserHandler)
	http.HandleFunc("/edit/", editUserHandler)
	http.HandleFunc("/save/", saveUserHandler)
	http.HandleFunc("/create/", displayCreateUserPageHandler)
	http.ListenAndServe(":8080", nil)
}
