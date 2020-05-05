package main

import (
	"io/ioutil"
	"log"
    "net/http"
    "html/template"
    "regexp"
    // "errors"
)

const root string  = "files/"
func mkT(name string) string {return "templates/" + name + ".html"}

var templates = template.Must(template.ParseFiles(mkT("edit"), mkT("view")))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var pageLink = regexp.MustCompile("[[a-zA-Z0-9]+]")

// Page Defines an article on the site
type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    filename := root + p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

// func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
//     m := validPath.FindStringSubmatch(r.URL.Path)
//     if m == nil {
//         http.NotFound(w, r)
//         return "", errors.New("invalid Page Title")
//     }
//     return m[2], nil // The title is the second subexpression.
// }

func loadPage(title string) (*Page, error) {
    filename := root + title + ".txt"
    body, err := ioutil.ReadFile(filename)

    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    // [] -> <a href="/view/PageName">PageName</a>
    // p.Body = pageLink.ReplaceAllFunc(p.Body, func(match []byte) []byte {
    //     // return []byte(`<a href="/view/Test">Test</a>`)
    //     return template.HTML(`<a href="/view/Test">Test</a>`)
    // })
    
    err := templates.ExecuteTemplate(w, tmpl+".html", p)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }

    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}
