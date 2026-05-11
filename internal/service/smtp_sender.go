package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"passworder/internal/model"
)

type MailSender interface {
	Send(settings model.MailSenderSettings, to string, subject string, body string) error
}

type SMTPSender struct{}

func NewSMTPSender() *SMTPSender {
	return &SMTPSender{}
}

func (s *SMTPSender) Send(settings model.MailSenderSettings, to string, subject string, body string) error {
	if settings.SMTPHost == "" || settings.SMTPPort == 0 || settings.SMTPFromAddress == "" {
		return fmt.Errorf("mail settings incomplete")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient email is empty")
	}

	fromName := strings.TrimSpace(settings.SMTPFromName)
	from := strings.TrimSpace(settings.SMTPFromAddress)
	transportFrom := from
	if strings.Contains(strings.TrimSpace(settings.SMTPUsername), "@") {
		transportFrom = strings.TrimSpace(settings.SMTPUsername)
	}
	if _, err := mail.ParseAddress(transportFrom); err != nil {
		return fmt.Errorf("invalid smtp sender address: %s", transportFrom)
	}
	fromHeader := transportFrom
	if _, err := mail.ParseAddress(from); err == nil && from != "" {
		fromHeader = from
	}
	if fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", fromName, transportFrom)
	}

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if strings.TrimSpace(settings.SMTPUsername) != "" {
		auth = smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	}

	if settings.SMTPPort == 465 {
		return sendMailWithTLS(settings, auth, transportFrom, to, []byte(message))
	}
	err := sendMailWithSTARTTLS(settings, auth, transportFrom, to, []byte(message))
	if err == nil || settings.SMTPPort != 587 {
		return err
	}

	fallbackSettings := settings
	fallbackSettings.SMTPPort = 465
	if fallbackErr := sendMailWithTLS(fallbackSettings, auth, transportFrom, to, []byte(message)); fallbackErr == nil {
		return nil
	}
	return err
}

func sendMailWithTLS(settings model.MailSenderSettings, auth smtp.Auth, from string, to string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: settings.SMTPHost})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Quit()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(message); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func sendMailWithSTARTTLS(settings model.MailSenderSettings, auth smtp.Auth, from string, to string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: settings.SMTPHost}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(message); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
