package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/fdelbos/mauth/generator"

	"github.com/fdelbos/mauth/sender/smtp"
	"github.com/fdelbos/mauth/templates/gotemplates"

	"github.com/fdelbos/mauth"
	"github.com/fdelbos/mauth/generator/hmac"
)

const (
	// the root url of the service
	baseURL  = "http://localhost:8080"
	htmlForm = `
<html>
<body>
	Enter your email address to receive a login link:<br/>
	<form method="post">
	<input type="text" name="email" />
	<input type="submit" value="login"/>
	</form><br/>
	You can check your emails in your <a href="http://localhost:8025/mail/">MailHog instance</a>
</body>
</html>
`
)

var (
	auth *mauth.MAuth
)

func init() {
	// A generator creates the token, here using HMAC
	generator, err := hmac.NewHMACB64("YRIXcOJ0wvjE14iJFz3eBCHj5qPzpq952LUcWlfPRDc=")
	if err != nil {
		log.Fatal(err)
	}

	// the sender sends the message
	sender, err := smtp.NewSMTP(smtp.Params{
		Host: "localhost",
		Port: 1025,
		From: "test@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	// the templates contains the email (optionally localized) templates to be send
	templates := gotemplates.NewTemplates()
	// one in english
	templates.Add("en", "hello!",
		`Follow the link to login: {{ .URL }}`,
		`Click the link to login: <a href="{{ .URL }}">{{ .URL }}</a>`)
	// one in french
	templates.Add("fr", "salut!",
		`Suivez le lien pour vous connecter: {{ .URL }}`,
		`Cliquez sur le lien pour vous connecter: <a href="{{ .URL }}">{{ .URL }}</a>`)

	auth, err = mauth.NewMAuth(generator, sender, templates, baseURL)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// a user posted an email address
		sendEmail(w, r)

	} else if r.Method == http.MethodGet {
		// maybe we have a token?
		token := r.URL.Query().Get(auth.Param)
		if token != "" {
			createCookieAndRedirect(w, r, token)
			return
		}

		// maybe we have a cookie already ?
		if cookie, err := r.Cookie("email"); err == nil {
			w.Write([]byte("Hello: " + cookie.Value))
			return
		}

		// then we print the form
		w.Write([]byte(htmlForm))

	} else {
		replyError(w, http.StatusMethodNotAllowed)

	}
}

func replyError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func sendEmail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		replyError(w, http.StatusBadRequest)
		return
	}
	email := r.PostFormValue("email")

	// if you want to send the email async, use the background context instead and a go routine
	err = auth.Send(r.Context(), email)
	if err != nil {
		replyError(w, http.StatusInternalServerError)
		return
	}
	w.Write([]byte("email sent to " + email))
}

func createCookieAndRedirect(w http.ResponseWriter, r *http.Request, token string) {
	email, err := auth.Validate(r.Context(), token)
	if err != nil {
		if errors.Is(err, generator.ErrInvalid) {
			// invalid token
			replyError(w, http.StatusBadRequest)
		} else {
			replyError(w, http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "email", Value: email})
	// we should redirect so that the token doesn't stay in the navbar...
	http.Redirect(w, r, baseURL, http.StatusTemporaryRedirect)
}

func main() {
	http.Handle("/", http.HandlerFunc(handler))
	http.ListenAndServe(":8080", nil)
}
