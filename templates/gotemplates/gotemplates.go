package gotemplates

import (
	"bytes"
	"errors"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	txtTemplate "text/template"
	"time"

	"github.com/fdelbos/mauth/templates"
	"golang.org/x/text/language"
)

type (
	locale struct {
		txt     *txtTemplate.Template
		html    *htmlTemplate.Template
		subject string
	}

	GoTemplates struct {
		locales map[language.Tag]locale
		//txt     map[language.Tag]*txtTemplate.Template
		//html    map[language.Tag]*htmlTemplate.Template
		tags    []language.Tag
		matcher language.Matcher
	}

	tmplType int
)

const (
	txtType tmplType = iota
	htmlType
	anyType
)

var (
	ErrUnsupportedLanguage   = errors.New("language is not supported")
	ErrNoTemplateForLanguage = errors.New("no template set for this language")
	ErrTemplatesEmpty        = errors.New("both templates are empty")
	ErrSubjectEmpty          = errors.New("subject is empty")
)

func NewTemplates() *GoTemplates {
	return &GoTemplates{
		tags:    []language.Tag{},
		locales: map[language.Tag]locale{},
	}
}

func (t *GoTemplates) hasTemplate(tag language.Tag, tmplType tmplType) bool {
	locale, ok := t.locales[tag]
	if !ok {
		return false
	}

	switch tmplType {
	case txtType:
		return locale.txt != nil
	case htmlType:
		return locale.html != nil
	default:
		return locale.txt != nil || locale.html != nil
	}
}

func (t *GoTemplates) SetDefaultLanguage(lang string) error {
	tag, err := language.Parse(lang)
	if err != nil {
		return ErrUnsupportedLanguage
	}

	if !t.hasTemplate(tag, anyType) {
		return ErrNoTemplateForLanguage
	}

	newTags := []language.Tag{tag}
	for _, t := range t.tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	t.tags = newTags
	t.matcher = language.NewMatcher(t.tags)
	return nil
}

func (t *GoTemplates) addLanguage(tag language.Tag) error {
	for _, t := range t.tags {
		if t == tag {
			return nil
		}
	}

	t.tags = append(t.tags, tag)
	t.matcher = language.NewMatcher(t.tags)
	return nil
}

func (t *GoTemplates) Add(lang, subject, txt, html string) error {
	if subject == "" {
		return ErrSubjectEmpty
	} else if txt == "" && html == "" {
		return ErrTemplatesEmpty
	}

	tag, err := language.Parse(lang)
	if err != nil {
		return ErrUnsupportedLanguage
	}

	locale := locale{subject: subject}

	if txt != "" {
		tmpl, err := txtTemplate.New("").Parse(txt)
		if err != nil {
			return err
		}
		locale.txt = tmpl
	}

	if html != "" {
		tmpl, err := htmlTemplate.New("").Parse(html)
		if err != nil {
			return err
		}
		locale.html = tmpl
	}

	t.locales[tag] = locale

	return t.addLanguage(tag)
}

func (t *GoTemplates) AddBytes(lang, subject string, txt, html []byte) error {
	return t.Add(lang, subject, string(txt), string(html))
}

func (t *GoTemplates) AddReader(lang, subject string, txt, html io.Reader) error {
	var err error

	var tmplTxt []byte
	if txt != nil {
		tmplTxt, err = ioutil.ReadAll(txt)
		if err != nil {
			return err
		}
	}

	var tmplHTML []byte
	if html != nil {
		tmplHTML, err = ioutil.ReadAll(html)
		if err != nil {
			return err
		}
	}
	return t.AddBytes(lang, subject, tmplTxt, tmplHTML)
}

func (t GoTemplates) generate(tag language.Tag, email, url string, expiration time.Time) (*templates.TemplateResult, error) {
	if !t.hasTemplate(tag, anyType) {
		return nil, ErrNoTemplateForLanguage
	}

	locale := t.locales[tag]

	res := templates.TemplateResult{Subject: locale.subject}

	if locale.txt != nil {
		dest := bytes.Buffer{}
		data := struct {
			Email      string
			URL        string
			Expiration time.Time
		}{Email: email, URL: url, Expiration: expiration}

		if err := locale.txt.Execute(&dest, data); err != nil {
			return nil, err
		}
		res.TXT = dest.Bytes()
	}

	if locale.html != nil {
		dest := bytes.Buffer{}
		data := struct {
			Email      string
			URL        string
			Expiration time.Time
		}{Email: email, URL: url, Expiration: expiration}

		if err := locale.html.Execute(&dest, data); err != nil {
			return nil, err
		}
		res.HTML = dest.Bytes()
	}

	return &res, nil
}

func (t GoTemplates) Generate(email, url string, expiration time.Time) (*templates.TemplateResult, error) {
	if t.tags == nil || len(t.tags) == 0 {
		return nil, ErrNoTemplateForLanguage
	}
	return t.generate(t.tags[0], email, url, expiration)
}

func (t GoTemplates) GenerateForLang(lang string, email, url string, expiration time.Time) (*templates.TemplateResult, error) {
	if t.tags == nil || len(t.tags) == 0 {
		return nil, ErrNoTemplateForLanguage
	}
	tag, err := language.Parse(lang)
	if err != nil {
		return t.generate(t.tags[0], email, url, expiration)
	}
	_, idx, _ := t.matcher.Match(tag)
	return t.generate(t.tags[idx], email, url, expiration)
}

func (t GoTemplates) GenerateForRequest(r *http.Request, email, url string, expiration time.Time) (*templates.TemplateResult, error) {
	if t.tags == nil || len(t.tags) == 0 {
		return nil, ErrNoTemplateForLanguage
	}
	tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil {
		return t.generate(t.tags[0], email, url, expiration)
	}
	_, idx, _ := t.matcher.Match(tags...)

	return t.generate(t.tags[idx], email, url, expiration)
}
