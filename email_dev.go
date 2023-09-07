package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/jordan-wright/email"
	"github.com/k3a/html2text"
)

func mainx() {
	// https://github.com/jordan-wright/email
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("~~ Testing Email Delivery")

	t, _ := template.ParseFiles("./emails/email-verify.html")
	var htmlBody bytes.Buffer

	t.Execute(&htmlBody, struct {
		Name    string
		Message string
	}{
		Name:    "Ron Londeau",
		Message: "Here ya go old buddy!",
	})

	plainBody := html2text.HTML2Text(htmlBody.String())

	e := email.NewEmail()
	e.From = "Hasura Base <hello@hasura-base.com>"
	e.To = []string{"ron@londeau.com"}
	e.Subject = "Email With Go!"
	e.Text = []byte(plainBody)
	e.HTML = htmlBody.Bytes()

	host := os.Getenv("AUTH_SMTP_HOST")
	port := os.Getenv("AUTH_SMTP_PORT")
	username := os.Getenv("AUTH_SMTP_USER")
	password := os.Getenv("AUTH_SMTP_PASS")

	sendError := e.Send(host+":"+port, smtp.PlainAuth("", username, password, host))
	if sendError != nil {
		log.Fatal(sendError)
	}

	fmt.Println("~~ Message Sent!")
}
