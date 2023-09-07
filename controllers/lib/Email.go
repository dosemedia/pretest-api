package lib

import (
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
	"github.com/k3a/html2text"
)

func SendEmail(from string, to string, subject string, htmlBody string) error {

	plainBody := html2text.HTML2Text(htmlBody)
	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(plainBody)
	e.HTML = []byte(htmlBody)

	host := os.Getenv("AUTH_SMTP_HOST")
	port := os.Getenv("AUTH_SMTP_PORT")
	username := os.Getenv("AUTH_SMTP_USER")
	password := os.Getenv("AUTH_SMTP_PASS")

	sendError := e.Send(host+":"+port, smtp.PlainAuth("", username, password, host))
	return sendError
}
