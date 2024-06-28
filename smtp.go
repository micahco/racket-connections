package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

func (app *application) sendMail(to, subject, html string) error {
	if !app.isProduction {
		fmt.Println("SMTP\tsendMail:", html)

		return nil
	}

	if err := app.smtpClient.Mail(app.fromAddress.Address); err != nil {
		return err
	}
	if err := app.smtpClient.Rcpt(to); err != nil {
		return err
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
		"\r\n<html><body>%s</body></html>\r\n",
		app.fromAddress.String(), to, subject, html)

	wc, err := app.smtpClient.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		return err
	}
	err = wc.Close()
	if err != nil {
		return err
	}

	err = app.smtpClient.Reset()
	if err != nil {
		return err
	}

	return nil
}

func newSMTPClient(host, port, user, pass string) (*smtp.Client, error) {
	smtpAddr := fmt.Sprintf("%s:%s", host, port)
	conn, err := tls.Dial("tcp", smtpAddr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	})
	if err != nil {
		return nil, err
	}

	sc, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}

	a := smtp.PlainAuth("", user, pass, host)
	if err := sc.Auth(a); err != nil {
		return nil, err
	}

	if err := sc.Noop(); err != nil {
		return nil, err
	}

	return sc, nil
}
