package models

import (
	"fmt"
	"net/smtp"
)

type MailConfig struct {
	Username  string
	Password  string
	Host      string
	Port      int64
	Recipient string
}

type EmailSender interface {
	Send(from string, body []byte) error
	Config() *MailConfig
}

type emailSender struct {
	conf MailConfig
	send func(string, smtp.Auth, string, []string, []byte) error
}

func (e *emailSender) Send(from string, body []byte) error {
	addr := fmt.Sprintf("%s:%d", e.conf.Host, e.conf.Port)
	auth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, e.conf.Host)
	return e.send(addr, auth, from, []string{e.conf.Recipient}, body)
}

func (e *emailSender) Config() *MailConfig {
	return &e.conf
}

func NewProductionSender(conf MailConfig) EmailSender {
	return &emailSender{conf, smtp.SendMail}
}

func NewSender(conf MailConfig, send func(string, smtp.Auth, string, []string, []byte) error) EmailSender {
	return &emailSender{conf, send}
}
