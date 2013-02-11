package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	// for reverse proxy redirects
	APP  = "/dat"
	ADDR = "localhost:8088"

	// url pattern
	ROOT = "/"

	// static resources like css and js
	S_URL = "/s/"
	S_DIR = "s"

	// special pages to display
	P_DIR  = "p"
	P_404  = "404"
	P_ROOT = "root"

	// convenience
	GET  = "GET"
	POST = "POST"

	// templates for pages, if you add one here, add it to ParseFiles below.
	//No constant slices allowed.
	//Also: templates referenced from inside of a template must be added.
	TMPLT_DIR    = "templates"
	ROOT_TMPLT   = "root"
	FOOTER_TMPLT = "footer"
)

var (
	templates = template.Must(template.ParseFiles(path.Join(TMPLT_DIR, ROOT_TMPLT), path.Join(TMPLT_DIR, FOOTER_TMPLT)))
)

type Page struct {
	// Content of the page, should be a <div id="main">...</div>
	Data  template.HTML
	title string
}

func loggerHandler(hf http.HandlerFunc) http.HandlerFunc {
	// Wrap a logger func around a request handler
	return func(w http.ResponseWriter, r *http.Request) {
		referer := r.Referer()
		logPattern := "%s - %s (%s)\n"
		switch r.Method {
		case GET:
			log.Printf(logPattern, GET, r.URL.Path, referer)
		case POST:
			log.Printf(logPattern, POST, r.URL.Path, referer)
		}
		hf(w, r)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	// Handles "/"
	var err error
	switch r.Method {
	case GET:
		// Load page content, which is the id="main"
		// XXX: rewrite to slurp the links for the sidebar from the P_DIR folder
		page := pageLoad(r.URL.Path[1:])
		// Show template of index
		err = render(w, ROOT_TMPLT, page)
	case POST:
		// Should not happen
		_, err = fmt.Fprint(w, "What are you doing?!")
	}
	if err != nil {
		handleErr(w, err)
		// Just in case more code added below
		return
	}
}

func pageLoad(page string) *Page {
	// XXX: This function sucks. Rewrite it.
	var b []byte
	var err error
	if page != "" {
		b, err = ioutil.ReadFile(path.Join(P_DIR, page))
		if err != nil {
			b, _ = ioutil.ReadFile(path.Join(P_DIR, P_404))
		}
	} else {
		b, _ = ioutil.ReadFile(path.Join(P_DIR, P_ROOT))
	}
	return &Page{Data: template.HTML(b), title: page}
}

func render(w http.ResponseWriter, name string, dot interface{}) error {
	err := templates.ExecuteTemplate(w, name, dot)
	if err != nil {
		return err
	}
	return nil
}

func handleErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
	// Initialize logger
	log.SetOutput(os.Stderr)
	// Handle request and wrap a logger around it
	http.HandleFunc(ROOT, loggerHandler(root))

	// Handle static file requests, no need to care about MIME types 
	http.Handle(S_URL, http.StripPrefix(S_URL, http.FileServer(http.Dir(S_DIR))))

	err := http.ListenAndServe(ADDR, nil)
	if err != nil {
		log.Panic(err)
	}
}
