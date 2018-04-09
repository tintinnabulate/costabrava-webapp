package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"io/ioutil"
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

// Data to put into template
type Page struct {
	Title string
	Body  string
}

// The template
var templateText string = `
<head>
  <title>{{.Title}}</title>
</head>

<body>
  {{.Body | markDown}}
</body>
`

// site configuration
var config configuration

// create a set of templates from many files.
var tmpls = template.Must(template.ParseGlob("*.tmpl"))

// the "getting here" page
var ghFile, _ = ioutil.ReadFile("getting_here.md")
var gettingHere = &Page{Title: "Getting here", Body: string(ghFile)}

// our markdown template
var markdownTmpl = template.Must(template.New("page.html").Funcs(template.FuncMap{"markDown": markDowner}).Parse(templateText))

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
	http.HandleFunc("/getting_here", markdownHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func markDowner(args ...interface{}) template.HTML {
	return template.HTML(blackfriday.Run([]byte(fmt.Sprintf("%s", args...))))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpls.ExecuteTemplate(w, "index", nil)
}

func markdownHandler(w http.ResponseWriter, r *http.Request) {
	err := markdownTmpl.ExecuteTemplate(w, "page.html", gettingHere)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}
