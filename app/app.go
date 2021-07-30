package app

import (
	"fmt"
	"net/http"
	"time"
)

type app struct {
	middleware      map[string]map[string][]func(http.ResponseWriter, *http.Request) (*http.Request, bool, []byte, int)
	routes          map[string][]string
	handlers        map[string]map[string]func(http.ResponseWriter, *http.Request)
	params          map[string]map[string][]string
	primeRoutes     map[string]map[string]bool
	secondaryRoutes map[string]map[string]bool
	viewsDir        string
	staticDir       string
	httpLoggerPath  string
	httpLoggerFile  string
}

type HttpRequestInfo struct {
	method    string
	uri       string
	referer   string
	ipaddr    string
	code      int
	size      int64
	duration  time.Duration
	userAgent string
}

func Initialize() *app {
	middleware := make(map[string]map[string][]func(http.ResponseWriter, *http.Request) (*http.Request, bool, []byte, int))
	params := make(map[string]map[string][]string)
	handlers := make(map[string]map[string]func(http.ResponseWriter, *http.Request))
	routes := make(map[string][]string)
	primeRoutes := make(map[string]map[string]bool)
	secondaryRoutes := make(map[string]map[string]bool)
	return &app{
		middleware:      middleware,
		routes:          routes,
		handlers:        handlers,
		params:          params,
		primeRoutes:     primeRoutes,
		secondaryRoutes: secondaryRoutes,
		viewsDir:        "./views/",
		staticDir:       "./static/",
		httpLoggerPath:  "./logs",
		httpLoggerFile:  "logs.txt",
	}
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	method := r.Method
	if _, ok := a.primeRoutes[method][url]; ok {
		a.ServeHttpPrimary(w, r, url, method)
	} else {
		a.ServeHttpSecondary(w, r, url, method)
	}
}

func (a *app) Listen(port int) {
	PORT := fmt.Sprintf(":%d", port)
	s := &http.Server{
		Addr:           PORT,
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("Server started on http://localhost:%d\n", port)
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (a *app) Set(key string, value string) {
	if key == "views" {
		a.viewsDir = value
	}
	if key == "static" {
		a.staticDir = value
	}
	if key == "log directory" {
		a.httpLoggerPath = value
	}
	if key == "log file" {
		a.httpLoggerFile = value
	}
}
