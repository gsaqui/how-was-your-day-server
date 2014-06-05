package main

import (
	"fmt"
	// "io/ioutil"
	"encoding/json"
	"flag"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	// "time"
)

type User struct {
	Name         string
	UserId       string
	UserStatuses []StatusUpdate
}

type StatusUpdate struct {
	UserId string
	State  int64
	Date   int64
}

type App struct {
	dbUrl               string
	channel             chan User
	statusUpdateChannel chan StatusUpdate
}

func Marshel(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

	s := string(b)
	return s
}

func UpdateUser(user User, dbUrl string) {
	c, err := redis.Dial("tcp", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	c.Send("SET", user.UserId, Marshel(user))
	c.Flush()
	c.Receive() // reply from SET
	defer c.Close()
}

func FindAllUsers(id string, dbUrl string) User {
	c, err := redis.Dial("tcp", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	c.Send("GET", id)
	c.Flush()
	v, err := redis.Bytes(c.Receive())
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	var m User
	err = json.Unmarshal(v, &m)

	return m
}

func (app App) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		if strings.ToLower(r.URL.Path) == "/newuser" {
			u, err := uuid.NewV4()
			if err != nil {
				log.Fatal(err)
			}
			value := r.FormValue("username")
			user := User{Name: value, UserId: u.String()}
			user.UserStatuses = make([]StatusUpdate, 0)

			app.channel <- user

			s := Marshel(user)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, s)
		}

		if strings.ToLower(r.URL.Path) == "/statusupdate" {

			state, err := strconv.ParseInt(r.FormValue("state"), 10, 64)
			if err != nil {
				state = 0
			}

			date, err := strconv.ParseInt(r.FormValue("date"), 10, 64)
			if err != nil {
				date = 0
			}

			statusUpdate := StatusUpdate{UserId: r.FormValue("userId"), State: state, Date: date}

			app.statusUpdateChannel <- statusUpdate
			s := Marshel(statusUpdate)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, s)
		}
	}

	if r.Method == "GET" {
		regularExpressions, _ := regexp.Compile("allupdates/([a-zA-Z0-9].*)")
		path := strings.ToLower(r.URL.Path)

		if regularExpressions.MatchString(path) {
			//we should really add some error checking here
			uuidOfUser := regularExpressions.FindStringSubmatch(path)[1]
			fmt.Println(uuidOfUser)

			tempUser3 := FindAllUsers(uuidOfUser, app.dbUrl)
			fmt.Println(tempUser3)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, tempUser3.Name)
		}

	}
}

func (app App) saveNewUser() {
	for {
		var user User
		user = <-app.channel
		UpdateUser(user, app.dbUrl)
	}
}

func (app App) updateUser() {
	for {
		var statusUpdate StatusUpdate
		statusUpdate = <-app.statusUpdateChannel

		c, err := redis.Dial("tcp", app.dbUrl)
		if err != nil {
			log.Fatal(err)
		}

		jsonEncodedUser, err := redis.Bytes(c.Do("GET", statusUpdate.UserId))
		if err != nil {
			log.Fatal(err)
		}

		var user User
		err = json.Unmarshal(jsonEncodedUser, &user)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(user.Name, user.UserStatuses)
		user.UserStatuses = append(user.UserStatuses, statusUpdate)
		fmt.Println(user.Name, user.UserStatuses)
		UpdateUser(user, app.dbUrl)
		defer c.Close()
	}
}

func main() {
	app := App{}

	app.channel = make(chan User)
	app.statusUpdateChannel = make(chan StatusUpdate)

	flag.StringVar(&app.dbUrl, "dbUrl", "http://notworking.com", "The url of the redis database")
	flag.Parse()

	go http.Handle("/", app)
	go app.saveNewUser()
	go app.updateUser()
	http.ListenAndServe("localhost:4000", nil)
}
