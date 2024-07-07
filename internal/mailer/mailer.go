package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"net/mail"
	"strconv"
	"text/template"

	"gopkg.in/gomail.v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender *mail.Address
}

func New(host, port, username, password string, sender *mail.Address) (*Mailer, error) {
	smtpPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	m := &Mailer{
		dialer: gomail.NewDialer(host, smtpPort, username, password),
		sender: sender,
	}
	return m, nil
}

func (m *Mailer) Send(recepient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("1: %s", err.Error())
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("2: %s", err.Error())
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return fmt.Errorf("3: %s", err.Error())
	}

	msg := gomail.NewMessage()
	msg.SetHeader("To", recepient)
	msg.SetHeader("From", m.sender.String())
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", body.String())

	return m.dialer.DialAndSend(msg)
}
