package smtp

import (
	"crypto/tls"
	"io"
	"io/ioutil"
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
		Sender     string
		From       string
		TimeOut    time.Duration
		TLSConfig  *tls.Config
		Logger     *log.Logger
	}

	SMTP struct {
		client *mail.SMTPClient
		sender string
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
		sender: params.Sender,
		from:   params.From,
		log:    params.Logger,
	}

	if res.log == nil {
		res.log = log.New(os.Stdout, "mauth smtp", log.LstdFlags)
	}

	return res, nil
}

func (s SMTP) Send(address, subject string, txt io.Reader, html io.Reader) error {
	email := mail.NewMSG()
	email.AddTo(address)
	if s.from != "" {
		email.SetFrom(s.from)
	}
	email.SetSender(s.sender)
	email.SetSubject(subject)

	if html != nil {
		body, err := ioutil.ReadAll(html)
		if err != nil {
			s.log.Printf("error while reading html body: '%s'", err)
			return err
		}
		email.SetBody(mail.TextHTML, string(body))

		if txt != nil {
			body, err := ioutil.ReadAll(txt)
			if err != nil {
				s.log.Printf("error while reading text body: '%s'", err)
				return err
			}
			email.AddAlternative(mail.TextPlain, string(body))
		}
	} else {
		body, err := ioutil.ReadAll(txt)
		if err != nil {
			s.log.Printf("error while reading text body: '%s'", err)
			return err
		}
		email.SetBody(mail.TextPlain, string(body))
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
