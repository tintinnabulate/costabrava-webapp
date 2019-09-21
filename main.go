package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/mail"
)

// create a set of templates from many files.
var tmpls = template.Must(template.ParseGlob("*.tmpl"))

const howToContactUs = `
Dear fellow member,

To contact the committee, please send your queries in an email to <info@costabravaconvention.com>.

We will do our best to get back to you in a timely manner.

In fellowship,

Costa Brava Committee.
`

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	registerHandlers()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func registerHandlers() {
	// define handlers
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "index", nil)
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		address := r.FormValue("email")
		ctx := appengine.NewContext(r)
		msg := &mail.Message{
			Sender:  "[DO NOT REPLY] Costa Brava Admin <donotreply@costabrava-1.appspotmail.com>",
			To:      []string{address},
			Subject: "How to Contact Us",
			Body:    howToContactUs,
		}
		if err := mail.Send(ctx, msg); err != nil {
			fmt.Errorf("Couldn't send email: %v", err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}
