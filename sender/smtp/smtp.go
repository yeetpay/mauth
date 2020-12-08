// Package smtp allow the sending of email message using the SMTP protocol.
package smtp

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type (
	encryption int
	auth       int

	Params struct {
		Host       string
		Port       int
		Username   string
		Password   string
		Encryption encryption
		Auth       auth
		From       string
		TimeOut    time.Duration
		TLSConfig  *tls.Config
		Logger     *log.Logger
	}

	SMTP struct {
		client *mail.SMTPClient
		from   string
		log    *log.Logger
	}
)

const (
	EncryptionNone encryption = iota
	EncryptionSSL
	EncryptionTLS

	AuthPlain auth = iota
	AuthLogin
	AuthCRAMMD5
)

func NewSMTP(params Params) (*SMTP, error) {
	server := mail.NewSMTPClient()
	server.Host = params.Host
	server.Port = params.Port
	server.Username = params.Username
	server.Password = params.Password

	switch params.Encryption {
	case EncryptionSSL:
		server.Encryption = mail.EncryptionSSL
	case EncryptionTLS:
		server.Encryption = mail.EncryptionTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	switch params.Auth {
	case AuthCRAMMD5:
		server.Authentication = mail.AuthCRAMMD5
	case AuthLogin:
		server.Authentication = mail.AuthLogin
	default:
		server.Authentication = mail.AuthPlain
	}

	if params.TimeOut != 0 {
		server.ConnectTimeout = params.TimeOut
		server.SendTimeout = params.TimeOut
	}

	if params.TLSConfig != nil {
		server.TLSConfig = params.TLSConfig
	}

	client, err := server.Connect()
	if err != nil {
		return nil, err
	}

	res := &SMTP{
		client: client,
		from:   params.From,
		log:    params.Logger,
	}

	if res.log == nil {
		res.log = log.New(os.Stdout, "mauth smtp", log.LstdFlags)
	}

	return res, nil
}

func (s SMTP) Send(ctx context.Context, address, subject string, txt, html []byte) error {
	email := mail.NewMSG()
	email.AddTo(address)
	email.SetFrom(s.from)

	email.SetSubject(subject)

	if html != nil {
		email.SetBody(mail.TextHTML, string(html))

		if txt != nil {
			email.AddAlternative(mail.TextPlain, string(txt))
		}
	} else {
		email.SetBody(mail.TextPlain, string(txt))
	}

	if err := email.Send(s.client); err != nil {
		s.log.Printf("error while sending the email: '%s'", err)
		return err
	}

	if err := email.GetError(); err != nil {
		s.log.Printf("error while sending the email: '%s'", err)
		return err
	}

	return nil
}
