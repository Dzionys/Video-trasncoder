package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"fmt"

	"./lp"
	"./sse"
	transcoder "./transcode"
)

var uploadtemplate = template.Must(template.ParseGlob("upload.html"))

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		w.WriteHeader(200)

		err := uploadtemplate.Execute(w, nil)
		if err != nil {
			log.Print(err)
		}
	case "POST":

		//Starts readig file by chuncking
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed to upload file")
			return
		}
		
		sse.UpdateFakeTerminalMessage(handler.Filename)
		// lp.WLog("Upload started")

		//Create empty file in /videos folder
		lp.WLog("Creating file")
		dst, err := os.OpenFile("./videos/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		defer dst.Close()
		if err != nil {
			log.Println(err)
			lp.WLog("Error: could not create file")
			return
		}

		//Copies a temporary file to empty file in /videos folder
		lp.WLog("Writing to file")
		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			lp.WLog("Error: failed to write file")
			return
		}

		lp.WLog("Upload successful")

		// Sending json response with video info
		var wg sync.WaitGroup
		wg.Add(1)
		infojs, err := transcoder.GetMediaInfoJson("./videos/"+handler.Filename, &wg)
		if err != nil {
			lp.WLog(fmt.Sprintf("%s", err))
		}
		w.Write(infojs)
		wg.Wait()

		// Start to transcode file.
		//go transcoder.ProcessVodFile(handler.Filename, true)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {

	// Make a new Broker instance
	sse.B = &sse.Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}

	// Start processing events
	sse.B.Start()

	http.Handle("/sse/dashboard", sse.B)
	http.Handle("/upload", http.HandlerFunc(uploadHandler))
	http.Handle("/", http.FileServer(http.Dir("views")))
	log.Println("Listening on port: 8080...")
	log.Fatalf("Exited: %s", http.ListenAndServe(":8080", nil))
}
