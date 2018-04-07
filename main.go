package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type configuration struct {
	SiteName     string
	SiteDomain   string
	SMTPServer   string
	SMTPUsername string
	SMTPPassword string
}

var config configuration
var tmpls = template.Must(template.ParseGlob("*.tmpl")) //create a set of templates from many files.

func init() {
	// load config
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config = configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}

	// define handlers
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpls.ExecuteTemplate(w, "index", nil)
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
