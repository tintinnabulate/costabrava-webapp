package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/tintinnabulate/gonfig"
)

var (
	config Config
	// create a set of templates from many files.
	tmpls = template.Must(template.ParseGlob("*.tmpl"))
)

const howToContactUs = `
Dear fellow member,

To contact the committee, please send your queries in an email to <%s>.

We will do our best to get back to you in a timely manner.

In fellowship,

Costa Brava Committee.
`

const howToContactUsHTML = `
<html>
<body>
<p>Dear fellow member,</p>

<p>To contact the committee, please send your queries in an email to
<a href="mailto:%s">%s</a>.</p>

<p>We will do our best to get back to you in a timely manner.</p>

<p>In fellowship,</p>

<p>Costa Brava Committee.</p>
</body>
</html>
`

func init() {
	configInit("config.json")
	handlersInit()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Start here: http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func ComposeEmail(address string) []byte {
	m := mail.NewV3Mail()

	sender_address := config.ContactEmail
	sender_name := "Costa Brava Convention Committee"
	e := mail.NewEmail(sender_name, sender_address)
	m.SetFrom(e)

	subject := "How to Contact Us"
	m.Subject = subject

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(address, address),
	}
	p.AddTos(tos...)
	m.AddPersonalizations(p)

	plainTextContent := fmt.Sprintf(howToContactUs, config.ContactEmail)
	htmlTextContent := fmt.Sprintf(howToContactUsHTML, config.ContactEmail, config.ContactEmail)
	c := mail.NewContent("text/plain", plainTextContent)
	m.AddContent(c)
	c = mail.NewContent("text/html", htmlTextContent)
	m.AddContent(c)

	mailSettings := mail.NewMailSettings()
	bypassListManagement := mail.NewSetting(true)
	mailSettings.SetBypassListManagement(bypassListManagement)
	spamCheckSetting := mail.NewSpamCheckSetting()
	spamCheckSetting.SetEnable(true)
	spamCheckSetting.SetSpamThreshold(1)
	spamCheckSetting.SetPostToURL("https://spamcatcher.sendgrid.com")
	mailSettings.SetSpamCheckSettings(spamCheckSetting)
	m.SetMailSettings(mailSettings)

	trackingSettings := mail.NewTrackingSettings()
	clickTrackingSettings := mail.NewClickTrackingSetting()
	clickTrackingSettings.SetEnable(true)
	clickTrackingSettings.SetEnableText(true)
	trackingSettings.SetClickTracking(clickTrackingSettings)
	openTrackingSetting := mail.NewOpenTrackingSetting()
	openTrackingSetting.SetEnable(true)
	trackingSettings.SetOpenTracking(openTrackingSetting)
	subscriptionTrackingSetting := mail.NewSubscriptionTrackingSetting()
	subscriptionTrackingSetting.SetEnable(true)
	trackingSettings.SetSubscriptionTracking(subscriptionTrackingSetting)
	m.SetTrackingSettings(trackingSettings)

	return mail.GetRequestBody(m)
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
			fmt.Fprintf(w, "Could not parse form: %v", err)
			return
		}

		address := r.FormValue("email")

		request := sendgrid.GetRequest(config.SendGridKey, "/v3/mail/send", "https://api.sendgrid.com")
		request.Method = "POST"
		request.Body = ComposeEmail(address)

		// send the email
		_, err := sendgrid.API(request)
		if err != nil {
			fmt.Fprintf(w, "Could not send email: %v", err)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func configInit(configName string) {
	err := gonfig.Load(&config, gonfig.Conf{
		FileDefaultFilename: configName,
		FileDecoder:         gonfig.DecoderJSON,
		FlagDisable:         true,
	})
	if err != nil {

		log.Fatalf("Could not load configuration file: %v", err)
	}
}

func handlersInit() {
	// define handlers
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

type Config struct {
	ContactEmail string `id:"ContactEmail"      default:"ContactEmail"`
	SendGridKey  string `id:"SendGridKey"       default:"SendGridKey"`
}
