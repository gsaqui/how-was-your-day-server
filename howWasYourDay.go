package main

import (
	"fmt"
	// "io/ioutil"
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
	"strconv"
	"strings"
	// "time"
)

type User struct {
	Name   string
	UserId string
}

type StatusUpdate struct {
	UserId string
	State  int64
}

type App struct {
}

func Marshel(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

	s := string(b)
	return s
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

			s := Marshel(user)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, s)
		}

		if strings.ToLower(r.URL.Path) == "/statusupdate" {
			value := r.FormValue("state")
			state, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				state = 0
			}

			statusUpdate := StatusUpdate{UserId: value, State: state}

			s := Marshel(statusUpdate)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, s)
		}

		//mongodb://user:secret3@ds045608.mongolab.com:45608/brainbrain

	}
}

func main() {
	app := App{}
	go http.Handle("/", app)
	http.ListenAndServe("localhost:4000", nil)
}
