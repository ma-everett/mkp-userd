
/* mkp-userd/examples/login/main.go */

package main

import (
	"log"
	"fmt"
	"net/http"
	"html/template"
	"time"
	"crypto/sha512"
	"encoding/hex"

	userd "../../client"
)

var (
	client *userd.Client
)


func main() {

	control := userd.NewControl(3 * time.Second)
	if err := control.Dial(); err != nil {
		log.Fatalf("unable to dial userd service - %v\n",err)
	}
	
	/* add alice */
	if _,err := control.Set(hash("alice","foobar1")); err != nil {

		log.Fatalf("unable to add alice - %v\n",err)
	} else {
		log.Printf("added alice with password foobar1\n")
	}

	/* add bob */
	if _,err := control.Set(hash("bob","barfoo")); err != nil {

		log.Fatalf("unable to add bob - %v\n",err)
	} else {
		log.Printf("added bob with password barfoo\n")
	}

	/* add sedwick */
	if _,err := control.Set(hash("sedwick","password")); err != nil {
		
		log.Fatalf("unable to add sedwick - %v\n",err)
	} else {
		log.Printf("added sedwick with password password\n")
	}
	
	/* attempt to fool the system */
	if _,err := control.Set(hash("sedwic","kpassword")); err != nil {

		log.Fatalf("unable to add sedwic - %v\n",err)
	} else {
		log.Printf("added sedwic with password kpassword\n")
	}

	control.Hangup()

	client = userd.NewClient(3 * time.Second,1 * time.Second)
	if err := client.Dial(); err != nil {
		log.Fatalf("unable to dial userd service - %v\n",err)
	}

	http.HandleFunc("/user/login/",loginHandler)
	http.HandleFunc("/user/home/",homeHandler)
	http.HandleFunc("/static/style.css",styleHandler)
	http.HandleFunc("/",rootHandler)

	log.Fatal(http.ListenAndServe("localhost:8080",nil))
}


var (
	templates = template.Must(template.ParseFiles("login.html"))
)

func render(w http.ResponseWriter,msg string) {
	err := templates.ExecuteTemplate(w,"login.html",msg)
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}
}

/* hash(hash(username) + password) */
func hash(username,password string) string {

	hash := sha512.New()
	hash.Write([]byte(username))
	h := hash.Sum(nil)

	hash = sha512.New()
	hash.Write([]byte(string(h) + password))
	h = hash.Sum(nil) 

	e := hex.EncodeToString(h)

	log.Printf("hash(hash(%s) + %s) = %s..\n",username,password,e[:8])
	return e
}

func loginHandler(w http.ResponseWriter, req *http.Request) {

	username := req.FormValue("username")
	password := req.FormValue("password")

	if len(username) <= 0 || len(password) <= 0 {

		log.Printf("login: invalid username or/and password\n")
		failedLoginHandler(w,req)
		return
	}

	h := hash(username,password)
	
	if ok,err := client.Check(h); err != nil {

		log.Printf("failed: userd is down\n")
		failedLoginHandler(w,req)
		return
	} else {

		if !ok { /* wrong username/password */
			log.Printf("login: user %s does not exist at userd\n",username)
			failedLoginHandler(w,req)
			return
		}
	}

	/* user found */
	http.Redirect(w,req,fmt.Sprintf("/user/home/?u=%s",username),http.StatusFound)	
}

func styleHandler(w http.ResponseWriter, req *http.Request) {

	http.ServeFile(w,req,"style.css")
}

func homeHandler(w http.ResponseWriter, req *http.Request) {

	render(w,"Home - FIXME")
}

func rootHandler(w http.ResponseWriter, req *http.Request) {

	render(w,"Please Login")
}

func failedLoginHandler(w http.ResponseWriter, req *http.Request) {

	render(w,"Unknown username and/or password, please try again.")
}
