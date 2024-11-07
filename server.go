package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
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

var templateError error
var errorTemplate *template.Template
var articleTemplate *template.Template
var homeTemplate *template.Template
var dashTemplate *template.Template

// note: has empty text field
var fileInfo []Article

var templates string = "./files/templates/"
var filePath string = "./files/static/"

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
	fileInfo = loadFileData()
	// TODO: need to configure a file server/handler to serve static files when needed from the server

	http.HandleFunc("/files/static/{file}", handleFile)
	http.HandleFunc("/{path}", handleHome)
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/article/{name}", handleRegArt)
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
	return articles
}

func handleFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("file")
	http.ServeFile(w, r, filePath+fileName)
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
func saveJson(fileName string, art Article) {
	data, err := json.Marshal(art)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("./articles/"+fileName+".json", data, 0777)
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
