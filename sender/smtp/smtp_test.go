package smtp_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/fdelbos/mauth/sender/smtp"
	"github.com/stretchr/testify/suite"
)

type (
	SMTPSuite struct {
		suite.Suite
		html     io.Reader
		text     io.Reader
		subject  string
		destAddr string
		fromAddr string
		sender   *smtp.SMTP
	}
)

var (
	htmlBody = []byte(`<div>html body</div>`)
	textBody = []byte(`text body`)
)

func TestSMTPSuite(t *testing.T) {
	suite.Run(t, &SMTPSuite{})
}

func (s *SMTPSuite) SetupTest() {
	s.subject = uniuri.New()
	s.destAddr = uniuri.NewLen(8) + "@dest.com"
	s.fromAddr = uniuri.NewLen(8) + "@sender.com"

	var err error
	s.sender, err = smtp.NewSMTP(smtp.Params{
		Host: "localhost",
		Port: 1025,
		From: s.fromAddr,
	})
	s.Require().Nil(err)
}

func (s *SMTPSuite) TestSendHTML() {
	err := s.sender.Send(
		s.destAddr,
		s.subject,
		nil,
		htmlBody)
	s.Require().Nil(err)
}

func (s *SMTPSuite) TestSendText() {
	err := s.sender.Send(
		s.destAddr,
		s.subject,
		textBody,
		nil)
	s.Require().Nil(err)
}

func (s *SMTPSuite) TestSendHTMLAndText() {
	err := s.sender.Send(
		s.destAddr,
		s.subject,
		textBody,
		htmlBody)
	s.Require().Nil(err)
	err = validateEmail(s.subject, s.destAddr, true, true)
	s.Require().Nil(err)
}

// TODO: finish it...
func validateEmail(subject, dest string, hasHTML, hasText bool) error {
	resp, err := http.Get("http://localhost:8025/mail/api/v2/search?kind=to&query=" + dest)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mailhog server returns http code %d", resp.StatusCode)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Print(string(res))
	return nil
}
