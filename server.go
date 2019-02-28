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
		defer file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		updateDashboard("Creating file")
		dst, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		defer dst.Close()
		if err != nil {
			fmt.Println(err)
			updateDashboard("Error while creating file")
			return
		}
		updateDashboard("Writing to file")
		if _, err := io.Copy(dst, file); err != nil {
			fmt.Println(err)
			updateDashboard("Error while writing to file")
			return
		} else {
			updateDashboard("Upload successful")
		}
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
	print(uploadlog + "\n")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	f, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("Streaming unsuported")
	}
	fmt.Fprintf(w, "data: %v\n\n", uploadlog)
	f.Flush()
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir("views")))
	http.HandleFunc("/sse/dashboard", dashboardHandler)
	log.Println("Listening on port: 8080...")
	log.Fatalf("Exited: %s", http.ListenAndServe(":8080", nil))
}
