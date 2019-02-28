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
		createLogMessage("Select video and upload")
		w.WriteHeader(200)
		//Executes upload file template
		err := uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}
	case "POST":
		createLogMessage("Upload started")
		//Starts readig file by chuncking
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("myfiles")
		defer file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		createLogMessage("Creating file")
		//Create empty file in /videos folder
		dst, err := os.OpenFile("./videos/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		defer dst.Close()
		if err != nil {
			fmt.Println(err)
			createLogMessage("Error while creating file")
			return
		}
		createLogMessage("Writing to file")
		//Copies a temporary file to empty file in /videos folder
		if _, err := io.Copy(dst, file); err != nil {
			fmt.Println(err)
			createLogMessage("Error while writing to file")
			return
		}
		createLogMessage("Upload successful")
		//Executes uploas template again after upload is complete
		err = uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//Creates log message
func createLogMessage(msg string) {
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
