package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"./lp"
	"./sse"
	transcoder "./transcode"
	"github.com/BurntSushi/toml"
)

var (
	uploadtemplate = template.Must(template.ParseGlob("upload.html"))
	vf             transcoder.Video
	wg             sync.WaitGroup
	crGot          = false
	CONF           transcoder.Config
)

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

		//lp.WLog("Upload started")

		// Checks if uploaded file with the same name already exists
		if _, err := os.Stat("./videos/" + handler.Filename); err == nil {
			lp.WLog(fmt.Sprintf("Error: file \"%v\" already exists", handler.Filename))
			return
		}

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

		// Load config file
		CONF, err := upConf()
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed to load config file")
			return
		}

		// Sending json response with video info
		data, err := transcoder.GetVidInfo("./videos/"+handler.Filename, CONF.TempJson, CONF.DataGen, CONF.TempTxt)
		if err != nil {
			log.Println(err)
			return
		}
		info, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed to marshal json file")
		}
		w.Write(info)

		// for {
		// 	if crGot {
		// 		break
		// 	}
		// }
		lp.WLog("Information received")

		// Start to transcode file.
		go transcoder.ProcessVodFile(handler.Filename, data, true)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func upConf() (transcoder.Config, error) {
	var conf transcoder.Config
	if _, err := toml.DecodeFile("transcode/conf.toml", &conf); err != nil {
		log.Println("error geting conf.toml: ", err)
		return conf, err
	}
	return conf, nil
}

func transcodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}

	// Decode json file
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&vf)
	if err != nil {
		log.Println(err)
		return
	}
	crGot = true
}

func main() {

	// Make a new Broker instance
	sse.B = &sse.Broker{
		Clients:        make(map[chan string]bool),
		NewClients:     make(chan (chan string)),
		DefunctClients: make(chan (chan string)),
		Messages:       make(chan string),
	}

	// Start processing events
	sse.B.Start()

	http.Handle("/transcode", http.HandlerFunc(transcodeHandler))
	http.Handle("/sse/dashboard", sse.B)
	http.Handle("/upload", http.HandlerFunc(uploadHandler))
	http.Handle("/", http.FileServer(http.Dir("views")))
	log.Println("Listening on port: 8080...")
	log.Fatalf("Exited: %s", http.ListenAndServe(":8080", nil))
}
