package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var uploadtemplate = template.Must(template.ParseGlob("./upload.html"))

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(200)
		err := uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}
	case "POST":
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("myfiles")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		z, err := os.OpenFile("./videos"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer z.Close()
		io.Copy(z, file)

		err = uploadtemplate.Execute(w, "Upload successful.")
		if err != nil {
			log.Print(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)
}
