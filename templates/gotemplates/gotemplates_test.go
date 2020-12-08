package gotemplates

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

var (
	htmlTemplates = map[string]map[string]string{
		"en": {
			"subject": "hello",
			"text":    `text: lang=en {{ .Email }} {{ .URL }} {{ .Expiration.Format "Jan 02, 2006"}}`,
			"html":    `<div>lang=en {{ .Email }} {{ .URL }} {{ .Expiration.Format "Jan 02, 2006"}}</div>`,
		},
		"fr": {
			"subject": "salut",
			"text":    `text: lang=fr {{ .Email }} {{ .URL }} {{ .Expiration.Format "Jan 02, 2006"}}`,
			"html":    `<div>lang=fr {{ .Email }} {{ .URL }} {{ .Expiration.Format "Jan 02, 2006"}}</div>`,
		},
	}

	expiration = time.Now()
	email      = "my.email@example.com"
	url        = uniuri.New()
)

func createTemplates(t *testing.T) *GoTemplates {
	tmpl := NewTemplates()
	for lang, group := range htmlTemplates {
		tmpl.Add(lang, group["subject"], group["text"], group["html"])
	}
	err := tmpl.SetDefaultLanguage("en")
	require.Nil(t, err)
	require.Equal(t, language.English, tmpl.tags[0])
	return tmpl
}

func TestCorrectTags(t *testing.T) {
	tmpl := createTemplates(t)
	require.Equal(t, tmpl.tags[0], language.English)
	require.Equal(t, tmpl.tags[1], language.French)
}

func TestSetDefaultLanguage(t *testing.T) {
	tmpl := createTemplates(t)
	err := tmpl.SetDefaultLanguage("fr")
	require.Nil(t, err)

	require.Equal(t, tmpl.tags[0], language.French)
	require.Equal(t, tmpl.tags[1], language.English)
}

func TestGenerate(t *testing.T) {
	tmpl := createTemplates(t)
	res, err := tmpl.Generate(email, url, expiration)
	require.Nil(t, err)
	require.Nil(t, validateTemplate("en", true, res.HTML))
	require.Nil(t, validateTemplate("en", false, res.TXT))
}

func TestGenerateForLang(t *testing.T) {
	tmpl := createTemplates(t)
	for _, al := range []struct {
		lang     string
		expected string
	}{
		{"en", "en"},
		{"en-US", "en"},
		{"fr", "fr"},
		{"zh", "en"},
		{"invalid!", "en"},
	} {
		res, err := tmpl.GenerateForLang(al.lang, email, url, expiration)
		require.Nil(t, err)
		require.Nil(t, validateTemplate(al.expected, true, res.HTML), al.lang)
		require.Nil(t, validateTemplate(al.expected, false, res.TXT), al.lang)
	}
}

func TestGenerateForRequest(t *testing.T) {
	tmpl := createTemplates(t)
	for _, al := range []struct {
		header   string
		expected string
	}{
		{"nn;q=0.3, en-us;q=0.8, en,", "en"},
		{"en-US,en;q=0.9,fr;q=0.8", "en"},
		{"fr-FR,en;q=0.8,fr;q=0.9", "fr"},
		{"gsw, en;q=0.7, en-US;q=0.8", "en"},
		{"gsw, nl, da", "en"},
		{"zh-ZH", "en"},
		{"invalid", "en"},
	} {
		r, _ := http.NewRequest("GET", "example.com", strings.NewReader("Hello"))
		r.Header.Set("Accept-Language", al.header)
		res, err := tmpl.GenerateForRequest(r, email, url, expiration)
		require.Nil(t, err, al.header)
		require.Nil(t, validateTemplate(al.expected, true, res.HTML), al.header, al.expected)
		require.Nil(t, validateTemplate(al.expected, false, res.TXT), al.header, al.expected)
	}
}

func validateTemplate(lang string, html bool, res []byte) error {
	if res == nil || len(res) == 0 {
		return errors.New("empty template")
	}
	str := string(res)
	if !strings.Contains(str, "lang="+lang) {
		return errors.New("invalid language " + string(res))
	}
	if html {
		if !strings.Contains(str, "<div>") {
			return errors.New("not html")
		}
	} else {
		if !strings.Contains(str, "text") {
			return errors.New("not text")
		}
	}

	if !strings.Contains(str, email) {
		return errors.New("email not present")
	}
	if !strings.Contains(str, url) {
		return errors.New("url not present")
	}
	if !strings.Contains(str, expiration.Format("Jan 02, 2006")) {
		return errors.New("expiration not present")
	}

	return nil
}
