package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "hza90plm"
	dbname   = "kyrsach"
)

type Article struct {
	Id       uint16
	Title    string
	Anons    string
	Text     string
	Category string
}

type Category struct {
	Id   uint16
	Name string
}

var posts []Article
var showPage Article
var categories []Category

func main() {
	handleFunc()
}

func handleFunc() {
	r := mux.NewRouter()
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/create/", create).Methods("GET")
	r.HandleFunc("/saveArticle/", saveArticle).Methods("POST")
	r.HandleFunc("/post/{id:[0-9]+}", showPost).Methods("GET")
	r.HandleFunc("/contacts/", showContacts).Methods("GET")

	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		return
	}
}

func showPost(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/show.html")
	vars := mux.Vars(r)

	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlConn)
	CheckError(err)

	defer db.Close()

	res, err := db.Query(fmt.Sprintf("SELECT id, title, anons, full_text FROM articles WHERE id = %s", vars["id"]))
	if err != nil {
		panic(err)
	}

	showPage = Article{}
	for res.Next() {
		var art Article
		err := res.Scan(&art.Id, &art.Title, &art.Anons, &art.Text)
		if err != nil {
			panic(err)
		}

		showPage = art
		fmt.Println(fmt.Sprintf("%s, %s, %s", art.Title, art.Anons, art.Text))
	}

	template.Execute(w, showPage)
}

func index(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/index.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlConn)
	CheckError(err)

	defer db.Close()

	res, err := db.Query("SELECT articles.id, title, anons, full_text, categoty_name FROM articles JOIN categories ON articles.category_id = categories.id")
	if err != nil {
		panic(err)
	}

	for res.Next() {
		var art Article
		err := res.Scan(&art.Id, &art.Title, &art.Anons, &art.Text, &art.Category)
		if err != nil {
			panic(err)
		}

		posts = append(posts, art)
		fmt.Println(fmt.Sprintf("%s, %s, %s, %s", art.Title, art.Anons, art.Text, art.Category))
	}

	template.Execute(w, posts)
	posts = nil
}

func create(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/create.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlConn)
	CheckError(err)

	defer db.Close()

	res, err := db.Query("SELECT categoty_name FROM categories")
	if err != nil {
		panic(err)
	}

	for res.Next() {
		var cat Category
		err := res.Scan(&cat.Name)
		if err != nil {
			panic(err)
		}

		categories = append(categories, cat)
	}
	fmt.Println(categories)
	template.Execute(w, categories)
	categories = nil
}

func showContacts(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/contacts.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	template.Execute(w, nil)
}

func saveArticle(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	text := r.FormValue("fullText")
	category := r.FormValue("category")
	fmt.Println(title, anons, text)

	if title == "" || anons == "" || text == "" {
		fmt.Fprintf(w, "Не все данные заполненые")
	} else {
		psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

		db, err := sql.Open("postgres", psqlConn)
		CheckError(err)

		defer db.Close()

		res, err := db.Query("SELECT id, categoty_name FROM categories")
		if err != nil {
			panic(err)
		}

		var categoryId uint16

		for res.Next() {
			var cat Category
			err := res.Scan(&cat.Id, &cat.Name)
			if err != nil {
				panic(err)
			}

			if category == cat.Name {
				categoryId = cat.Id
			}
		}

		sqlCode := `INSERT INTO articles (title, anons, full_text, category_id) VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(sqlCode, title, anons, text, categoryId)
		CheckError(err)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
