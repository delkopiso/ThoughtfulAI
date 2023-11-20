package views

import (
	"embed"
	"html/template"
	"path"
	"time"
)

const layout = "layout.gohtml"

//go:embed *
var files embed.FS

func parsePartial(file string) *template.Template {
	name := path.Base(file)
	return template.Must(template.New(name).ParseFS(files, file))
}

func parseTemplate(file string) *template.Template {
	return template.Must(template.New(layout).ParseFS(files, layout, file))
}

var (
	Index            = parseTemplate("templates/index.gohtml")
	Login            = parseTemplate("templates/login.gohtml")
	Watchlist        = parsePartial("partials/watchlist.gohtml")
	SearchSecurities = parsePartial("partials/search_securities.gohtml")
)

type IndexParams struct {
	CsrfName    string
	CsrfToken   string
	CurrentUser *User
}

type LoginParams struct {
	CsrfName        string
	CsrfToken       string
	CurrentUser     *User
	Username        string
	ValidationError string
}

type WatchlistParams struct {
	CsrfName    string
	CsrfToken   string
	CurrentUser *User
	Watchlist   []Security
}

type SearchSecuritiesParams struct {
	CsrfName   string
	CsrfToken  string
	Message    string
	Securities []Security
}

type Security struct {
	Ticker           string     `db:"ticker"`
	Name             string     `db:"name"`
	LastPrice        *float64   `db:"last_price"`
	LastPriceUpdated *time.Time `db:"last_price_updated"`
	Watched          bool       `db:"watched"`
}

type User struct {
	Id        string  `db:"user_id"`
	FirstName *string `db:"first_name"`
	LastName  *string `db:"last_name"`
	Email     string  `db:"email"`
}

func (u User) DisplayName() string {
	if u.FirstName != nil && u.LastName != nil {
		return *u.FirstName + " " + *u.LastName
	}
	return u.Email
}
