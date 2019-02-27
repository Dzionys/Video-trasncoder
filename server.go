package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	uploadlog      string
	uploadtemplate = template.Must(template.ParseGlob("upload.html"))
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		updateDashboard("Select video and upload")
		w.WriteHeader(200)
		err := uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}
	case "POST":
		updateDashboard("Upload started")
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("myfiles")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		updateDashboard("Creating file")
		z, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}

		updateDashboard("Writing to file")
		defer z.Close()
		io.Copy(z, file)
		updateDashboard("Upload successful")

		err = uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func updateDashboard(msg string) {
	curentTime := time.Now().Format("15:04:05")
	uploadlog = fmt.Sprintf("<%v> %v", curentTime, msg)
	print(uploadlog)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "data: %v\n\n", uploadlog)
	fmt.Printf("data: %v\n", uploadlog)
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir("views")))
	http.HandleFunc("/sse/dashboard", dashboardHandler)
	http.ListenAndServe(":8080", nil)
}
