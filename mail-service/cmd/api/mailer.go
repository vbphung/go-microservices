package main

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	fmtMsg, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMsg, err := m.buildPlainMessage(msg)
	if err != nil {
		return err
	}

	svr := mail.NewSMTPClient()
	svr.Host = m.Host
	svr.Port = m.Port
	svr.Username = m.Username
	svr.Password = m.Password
	svr.Encryption = m.getEncryption()
	svr.KeepAlive = false
	svr.ConnectTimeout = 10 * time.Second
	svr.SendTimeout = 10 * time.Second

	smtpClient, err := svr.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMsg)
	email.AddAlternative(mail.TextHTML, fmtMsg)

	if len(msg.Attachments) > 0 {
		for _, at := range msg.Attachments {
			email.AddAttachment(at)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) getEncryption() mail.Encryption {
	switch m.Encryption {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) buildPlainMessage(msg Message) (string, error) {
	renderTmpl := "./templates/mail.plain.gohtml"

	tmpl, err := template.New("email-plain").ParseFiles(renderTmpl)
	if err != nil {
		return "", err
	}

	var resTmpl bytes.Buffer
	if err = tmpl.ExecuteTemplate(&resTmpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	return resTmpl.String(), nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	renderTmpl := "./templates/mail.html.gohtml"

	tmpl, err := template.New("email-html").ParseFiles(renderTmpl)
	if err != nil {
		return "", err
	}

	var resTmpl bytes.Buffer
	if err = tmpl.ExecuteTemplate(&resTmpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	fmtMsg := resTmpl.String()

	return m.inlineCSS(fmtMsg)
}

func (m *Mail) inlineCSS(msg string) (string, error) {
	opts := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(msg, &opts)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
