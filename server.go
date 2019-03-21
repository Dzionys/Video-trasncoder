package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	crGot          = 0
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

		lp.WLog("Upload started")

		//Starts readig file by chuncking
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed to upload file")
			return
		}

		// Check if video file format is allowed
		allowed := false
		for _, ave := range CONF.FileTypes {
			if filepath.Ext(handler.Filename) == ave {
				allowed = true
			}
		}
		if !allowed {
			lp.WLog("Error: this file format is not allowed " + filepath.Ext(handler.Filename))
			return
		}

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
			removeFile("./videos/" + handler.Filename)
			return
		}
		lp.WLog("Upload successful")

		data, err := writeJsonResponse(w, handler.Filename)
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed send video data to client")
			removeFile("./videos/" + handler.Filename)
			return
		}

		sse.UpdateMessage(handler.Filename)

		if CONF.Advanced {
			go waitForClientData(handler.Filename, data)
		} else {
			go transcoder.ProcessVodFile(handler.Filename, data, vf)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Waits for client to send instructions about file transcoding
func waitForClientData(filename string, data transcoder.Vidinfo) {
	for {
		if crGot == 1 {
			lp.WLog("Information received")

			// Strart transcoding if data is received
			transcoder.ProcessVodFile(filename, data, vf)
			break
		} else if crGot == 2 {
			lp.WLog("Error: failed receiving information from client")
			removeFile("./videos/" + filename)
			return
		}
	}
}

// Send json response after file upload
func writeJsonResponse(w http.ResponseWriter, filename string) (transcoder.Vidinfo, error) {

	data, err := transcoder.GetVidInfo("./videos/"+filename, CONF.TempJson, CONF.DataGen, CONF.TempTxt)
	if err != nil {
		log.Println(err)
		return data, err
	}
	info, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return data, err
	}
	w.WriteHeader(200)
	w.Write(info)

	return data, nil
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
		crGot = 2
		log.Println(err)
		return
	}

	// Decode json file
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&vf)
	if err != nil {
		crGot = 2
		log.Println(err)
		return
	}
	crGot = 1
}

func removeFile(filepath string) {
	if os.Remove(filepath) != nil {
		lp.WLog("Error: failed removing source file")
	}
	return
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

	// Load config file
	var err error
	CONF, err = upConf()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to load config file")
		return
	}

	// Write all logs to file
	err = lp.OpenLogFile(CONF.LogP)
	if err != nil {
		lp.WLog("Error: failed open log file")
		return
	}
	defer lp.LogFile.Close()

	http.Handle("/transcode", http.HandlerFunc(transcodeHandler))
	http.Handle("/sse/dashboard", sse.B)
	http.Handle("/upload", http.HandlerFunc(uploadHandler))
	http.Handle("/", http.FileServer(http.Dir("views")))
	fmt.Println("Listening on port: 8080...")
	log.Fatalf("Exited: %s", http.ListenAndServe(":8080", nil))
}
