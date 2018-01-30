package main

import (
	_"fmt"
	"net/http"
    "html/template"

    "github.com/gorilla/sessions"
    "github.com/gorilla/mux"
    "github.com/go-redis/redis"
    "golang.org/x/crypto/bcrypt"
)

var client *redis.Client
var store = sessions.NewCookieStore([]byte("t0p-s3cr3taerg"))
var templates *template.Template

func main() {
    client = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    templates = template.Must(template.ParseGlob("template/*.html"))
    r := mux.NewRouter()
    r.HandleFunc("/", gethandler).Methods("GET")
    r.HandleFunc("/", posthandler).Methods("POST")
    r.HandleFunc("/login", logingethandler).Methods("GET")
    r.HandleFunc("/login", loginposthandler).Methods("POST")
    r.HandleFunc("/register", registergethandler).Methods("GET")
    r.HandleFunc("/register", registerposthandler).Methods("POST")
 // r.HandleFunc("/test", testgethandler).Methods("GET")
    fs := http.FileServer(http.Dir("./static/"))
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
    http.Handle("/", r)
    http.ListenAndServe(":8000", nil)
}

func gethandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    _, ok := session.Values["username"]
    if !ok {
        http.Redirect(w, r, "/login", 302)
        return
    }

    comments, err := client.LRange("comments", 0, 10).Result()
    if err != nil {
        return
    }
    templates.ExecuteTemplate(w,"index.html", comments)
}

func posthandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    comment := r.PostForm.Get("comment")
    client.LPush("comments", comment)
    http.Redirect(w, r, "/", 302)
}

func logingethandler(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w,"login.html", nil)
}

func loginposthandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    username := r.PostForm.Get("username")
    password := r.PostForm.Get("password")
    hash, err := client.Get("user:"+ username).Bytes()
    if err != nil {
        return
    }
    err = bcrypt.CompareHashAndPassword(hash, []byte(password))
    if err != nil {
        return
    }
    session, _ := store.Get(r, "session")
    session.Values["username"] = username
    session.Save(r,w)
    http.Redirect(w, r, "/", 302)
}

func registergethandler(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w,"register.html", nil)
}

func registerposthandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    username := r.PostForm.Get("username")
    password := r.PostForm.Get("password")
    cost := bcrypt.DefaultCost
    hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    if err != nil {
        return
    }
    client.Set("user:" + username, hash, 0)
    http.Redirect(w, r, "/login", 302)
}

//func testgethandler(w http.ResponseWriter, r *http.Request) {
//    session, _ := store.Get(r, "session")
//    untyped, ok := session.Values["username"]
//    if !ok {
//        return
//    }
//    username, ok := untyped.(string)
//    if !ok {
//        return
//    }
//    w.Write([]byte(username))
//}
