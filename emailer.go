package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
)

type request struct {
	from    mail.Address
	to      mail.Address
	subject string
	body    string
}

func newRequest(to, subject, body string) *request {
	return &request{
		from:    mail.Address{Name: "", Address: config.SMTPUsername},
		to:      mail.Address{Name: "", Address: to},
		subject: subject,
		body:    body,
	}
}

func emailCode(recipient, code string) {
	server := config.SMTPServer
	subj := fmt.Sprintf("%s Verification Code", config.SiteName)
	body := fmt.Sprintf("Verification link: <a href=\"%s/verify/%s\">Verify email address</a>", config.SiteDomain, code)
	r := newRequest(recipient, subj, body)
	r.sendEmail(server)
}

func (r *request) makeEmailHeaders() map[string]string {
	headers := make(map[string]string)
	headers["From"] = r.from.String()
	headers["To"] = r.to.String()
	headers["Subject"] = r.subject
	return headers
}

func (r *request) makeEmailMessage() string {
	message := ""
	for k, v := range r.makeEmailHeaders() {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + r.body
	return message
}

func (r *request) sendEmail(server string) {
	message := r.makeEmailMessage()

	// Connect to the SMTP Server
	serverName := fmt.Sprintf("%s:465", server)
	host, _, _ := net.SplitHostPort(serverName)
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPServer)

	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// For smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", serverName, tlsConfig)
	if err != nil {
		log.Panic(err)
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}
	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}
	// To && From
	if err = c.Mail(r.from.Address); err != nil {
		log.Panic(err)
	}
	if err = c.Rcpt(r.to.Address); err != nil {
		log.Panic(err)
	}
	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}
	err = w.Close()
	if err != nil {
		log.Panic(err)
	}
	c.Quit()
}
