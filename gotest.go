package main

import (
  _ "github.com/mattn/go-sqlite3"
  "database/sql"
  "io/ioutil"
  "net/http"
  "html/template"
  "regexp"
  "errors"
  "fmt"
  "os"
)
const lenPath = len("/view/")
const rootDir = "/Users/enzo/workspace/go/src/github.com/vincenzoauteri/gotest/"

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")


type Page struct {
  Title string
  Body []byte
}

var templates *template.Template


func parseTemplates(templateDir string) {
  templates = template.Must(template.ParseGlob(templateDir + "*html"))
}

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err error) {
  title = r.URL.Path[lenPath:]
  if !titleValidator.MatchString(title) {
    http.NotFound(w, r)
    err = errors.New("Invalid Page Title")
  }
  return
}

func (p *Page) save() error {
  const dataDir = "data/"
  filename := rootDir + dataDir + p.Title + ".txt"
  return ioutil.WriteFile(filename,p.Body,0600)
}

func loadPage(title string) (*Page, error) {
  const dataDir = "data/"
  filename := rootDir + dataDir + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if (err != nil) {
    return nil, err
  }
  return &Page{Title:title,Body:body},nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, "renderTemplate:" + err.Error(), http.StatusInternalServerError)
    return
  }
}

func makeHandler (fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[lenPath:]
    if !titleValidator.MatchString(title) {
      http.NotFound(w, r)
      return
    }
  fn (w, r, title)
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

func editHandler(w http.ResponseWriter, r *http.Request, title string ) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler (w http.ResponseWriter, r *http.Request, title string) { 
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, "saveHandler:" + err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w,r,"/view/"+title,http.StatusFound)
}

func initDb (dbName string) (*sql.DB, error) {
  os.Remove(rootDir + "data/"+ dbName + ".db")
  db, err := sql.Open("sqlite3", rootDir + "data/"+ dbName + ".db")
  if err != nil {
    fmt.Println(err)
    return nil,err
  }
  defer db.Close()
  sqls := []string{
    "create table foo (id integer not null primary key, name text)",
    "delete from foo",
  }

  for _, sql := range sqls {
    _, err = db.Exec(sql)
    if err != nil {
      fmt.Printf("%q: %s\n", err, sql)
      return nil,err
    }
  }
  return db,nil
}

func main () {
  fmt.Println("Staring Webserver at 8080")
  parseTemplates(rootDir + "templates/")
  _ ,err := initDb("foo") 
  if err != nil {
    fmt.Println(err)
    return
  }

  http.HandleFunc("/view/",makeHandler(viewHandler))
  http.HandleFunc("/edit/",makeHandler(editHandler))
  http.HandleFunc("/save/",makeHandler(saveHandler))
  http.ListenAndServe(":8080",nil)
}
