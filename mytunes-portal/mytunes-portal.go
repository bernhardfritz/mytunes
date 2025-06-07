package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bernhardfritz/mytunes/mytunes-portal/internal"
)

type Context struct {
	tss       *internal.TransientSessionStorage
	indexHtml *template.Template
}

func NewContext(tss *internal.TransientSessionStorage, indexHtml *template.Template) *Context {
	return &Context{
		tss:       tss,
		indexHtml: indexHtml,
	}
}

func loggingHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req.Method, req.RemoteAddr, req.URL.Path)
		handler.ServeHTTP(res, req)
	})
}

type Page struct {
	Host    string
	Token   string
	Android bool
	Chrome  bool
}

func (ctx *Context) handleRoot(res http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("_forward_auth")
	if err != nil {
		log.Println(err)
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	cookieToken, err := ctx.tss.StoreCookie(cookie.String())
	if err != nil {
		log.Println(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	userAgent := req.Header.Get("User-Agent")
	page := Page{
		Host:    req.Header.Get("X-Forwarded-Host"),
		Token:   cookieToken,
		Android: strings.Contains(userAgent, "Android"),
		Chrome:  strings.Contains(userAgent, "Chrome"),
	}

	err = ctx.indexHtml.Execute(res, page)
	if err != nil {
		log.Println(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
	}
}

func (ctx *Context) handleVlc(res http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	token := query.Get("token")
	if token == "" {
		http.Error(res, "Bad request", http.StatusBadRequest)
		return
	}

	cookieString, err := ctx.tss.FindCookie(token)
	if err != nil {
		log.Println(err)
		http.Error(res, "Not authorized", http.StatusUnauthorized)
		return
	}

	err = ctx.tss.DeleteCookie(token)
	if err != nil {
		log.Println(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Add("Set-Cookie", cookieString)

	http.Redirect(res, req, "/index.m3u", http.StatusTemporaryRedirect)
}

//go:embed index.html
var indexHtml string

func main() {
	mytunesKey := os.Getenv("MYTUNES_PORTAL_KEY")
	if mytunesKey == "" {
		log.Fatal("The MYTUNES_PORTAL_KEY environment variable is empty or not set.")
	}

	encde, err := internal.NewEncde([]byte(mytunesKey))
	if err != nil {
		log.Fatal(err)
	}

	tss, err := internal.NewTransientSessionStorage(encde)
	if err != nil {
		log.Fatal(err)
	}
	defer tss.Close()
	err = tss.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("index.html").Parse(indexHtml)
	if err != nil {
		log.Fatal(err)
	}

	ctx := NewContext(tss, tmpl)

	http.Handle("GET /{$}", loggingHandler(http.HandlerFunc(ctx.handleRoot)))
	http.Handle("GET /_vlc", loggingHandler(http.HandlerFunc(ctx.handleVlc)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
