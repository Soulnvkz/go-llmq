package endpoints

import (
	"log"
	"net/http"
	"text/template"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	// io.WriteString(w, "This is root!\n")

	var tmplFile = "templates/main.html"
	tmpl, err := template.New("main.html").ParseFiles(tmplFile)
	if err != nil {
		log.Default().Println("error:", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Default().Println("error:", err)
	}
}
