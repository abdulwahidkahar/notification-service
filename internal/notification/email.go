package notification

import (
	"fmt"
	"os"
	"strconv"

	gomail "gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func NewEmailConfig() (*EmailConfig, error) {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return &EmailConfig{}, err
	}
	return &EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}, nil
}

func (e *EmailConfig) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.Host, e.Port, e.User, e.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("gagal kirim email: %w", err)
	}

	return nil
}
