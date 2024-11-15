package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"slices"
	"time"
)

type Article struct {
	Title       string
	Author      string
	Description string
	Text        string
	Created     time.Time
	Updated     time.Time
}

const url string = ":8080"

const admin_uname string = "admin"
const admin_pass string = "admin"

var templateError error
var errorTemplate *template.Template
var articleTemplate *template.Template
var homeTemplate *template.Template
var dashTemplate *template.Template
var successTemplate *template.Template
var formTemplate *template.Template

// note: has empty text field
var fileInfo []Article

var templates string = "./files/templates/"

func main() {
	errorTemplate, templateError = template.ParseFiles(templates + "error.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}
	articleTemplate, templateError = template.ParseFiles(templates + "article.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}
	homeTemplate, templateError = template.ParseFiles(templates + "home.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}
	dashTemplate, templateError = template.ParseFiles(templates + "dash.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}
	successTemplate, templateError = template.ParseFiles(templates + "success.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}
	formTemplate, templateError = template.ParseFiles(templates + "form.template.html")
	if templateError != nil {
		log.Fatal(templateError)
	}

	fileInfo = loadFileData()

	http.HandleFunc("POST /new", handleCreate)
	http.HandleFunc("GET /new", handleNew)
	http.HandleFunc("GET /edit", handleEdit)
	http.HandleFunc("POST /edit", handleEditFile)
	http.HandleFunc("GET /admin", handleAdmin)
	http.Handle("GET /files/", http.StripPrefix("/files", http.FileServer(http.Dir("./files/static"))))
	http.HandleFunc("GET /{path...}", handleHome)
	http.HandleFunc("GET /article/{name}", handleRegArt)
	log.Fatal(http.ListenAndServe(url, nil))
}

func loadFileData() []Article {
	files, err := os.ReadDir("./articles")
	if err != nil {
		log.Fatal(err)
	}
	articles := make([]Article, 0, 5)
	for _, file := range files {
		bytes, err := os.ReadFile("./articles/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		// unmarshall all json
		var art Article
		json.Unmarshal(bytes, &art)
		art.Text = ""
		articles = append(articles, art)
	}
	slices.SortFunc(articles, func(a, b Article) int {
		return a.Created.Compare(b.Created)
	})
	return articles
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	err := dashTemplate.Execute(w, fileInfo)
	if err != nil {
		log.Fatal(err)
	}
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.PostFormValue("title")
	author := r.PostFormValue("author")
	text := r.PostFormValue("text")
	if title == "" {
		errorPage(w, errors.New("title is invalid"))
		return
	}
	_, err := os.Stat("./articles/" + title + ".json")
	if errors.Is(err, os.ErrNotExist) {
		var art Article = Article{Title: title, Author: author, Text: text, Created: time.Now(), Updated: time.Now()}
		fileInfo = append(fileInfo, art)
		saveJson(art)
	} else {
		errorPage(w, errors.New("article with same name already exists"))
		return
	}
	successPage(w, "Successfully created article")
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	uname, pass, ok := r.BasicAuth()
	if !ok {
		// send pass request header
		w.Header().Set("www-Authenticate", "Basic realm=\"User Visible Realm\"")
		w.WriteHeader(401)
		return
	}
	if uname == admin_uname && pass == admin_pass {
		// serve file
		formTemplate.Execute(w, "new")
		return
	}
	w.WriteHeader(404)
}

// TODO: add basic auth here as well
func handleEdit(w http.ResponseWriter, r *http.Request) {
	// check if there is a file with the selected name
	fileName := r.URL.Query().Get("file")
	// fetch file and fill in information into template form page
	_, err := getArticle(fileName)
	// if it doesn't exist then return error page
	// could also serve with a front-end script to fill in values
	if err != nil {
		errorPage(w, err)
		return
	}

	formTemplate.Execute(w, "edit")
	// serve page
}

func handleEditFile(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("helo ther"))
}

// handles request for homepage
func handleHome(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("path")
	switch name {
	case "home":
		{
			getRegHome(w)
		}
	case "":
		{
			getRegHome(w)
		}
	default:
		{
			errorPage(w, errors.New("no such page exists"))
		}
	}
}

// handles request to view article
func handleRegArt(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	getFileHtml(name, w)
}

// saves json value to file
func saveJson(art Article) {
	data, err := json.Marshal(art)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("./articles/"+art.Title+".json", data, 0777)
}

// sends home template to client
// should rework this to use stored map values
func getRegHome(w http.ResponseWriter) {
	err := homeTemplate.Execute(w, fileInfo)
	if err != nil {
		log.Fatal(err)
	}
}

// returns parsed json in article format from file
func getArticle(fileName string) (Article, error) {
	file, err := os.ReadFile("./articles/" + fileName + ".json")
	// need to handle file not existing here
	var art Article
	if err != nil {
		return art, err
	}
	err = json.Unmarshal(file, &art)
	if err != nil {
		return art, err
	}
	return art, nil
}

// sends client templated file data
func getFileHtml(name string, wr http.ResponseWriter) {
	art, err := getArticle(name)
	if err != nil {
		errorPage(wr, err)
		return
	}
	err = articleTemplate.Execute(wr, art)
	if err != nil {
		log.Fatal(err)
	}
}

// sends client an error page
func errorPage(w http.ResponseWriter, err error) {
	err = errorTemplate.Execute(w, err)
	if err != nil {
		log.Fatal(err)
	}
}

func successPage(w http.ResponseWriter, data string) {
	err := successTemplate.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}
