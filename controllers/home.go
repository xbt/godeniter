package controllers

import (
	"html/template"
	"net/http"
)

type HomeController struct{}

func (c *HomeController) Index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/home/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (c *HomeController) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Godeniter!"))
}
