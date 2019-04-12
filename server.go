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

	cf "./conf"
	db "./database"
	"./lp"
	tc "./transcode"
	vd "./videodata"
)

var (
	uploadtemplate string
	basetemplate   = "./views/templates/base.html"
	vf             vd.Video
	prd            vd.PData
	wg             sync.WaitGroup
	crGot          = 0
	CONF           cf.Config
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		w.WriteHeader(200)

		t, err := template.ParseFiles(basetemplate, uploadtemplate)
		if err != nil {
			log.Panicln(err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Panicln(err)
		}
	case "POST":

		lp.WLog("Upload started")

		//Starts readig file by chuncking
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			lp.WLog("Error: failed to upload file")
			return
		}
		defer file.Close()

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
			removeFile("./videos/", handler.Filename)
			return
		}
		lp.WLog("Upload successful")

		data, err := writeJsonResponse(w, handler.Filename)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			lp.WLog("Error: failed send video data to client")
			removeFile("./videos/", handler.Filename)
			return
		}

		err = db.InsertVideo(data, handler.Filename, "Preparing", -1)
		if err != nil {
			lp.WLog("Error: failed to insert video data in database")
			log.Println(err)
			removeFile("./videos/", handler.Filename)
			return
		}

		lp.UpdateMessage(handler.Filename)

		if CONF.Advanced {
			go waitForClientData(handler.Filename, data)
		} else {
			go tc.ProcessVodFile(handler.Filename, data, vf, prd, CONF)
			resetData()
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Waits for client to send instructions about file transcoding
func waitForClientData(filename string, data vd.Vidinfo) {
	for {
		if crGot == 1 {
			lp.WLog("Information received")

			// Strart transcoding if data is received
			go tc.ProcessVodFile(filename, data, vf, prd, CONF)
			resetData()
			break
		} else if crGot == 2 {
			lp.WLog("Error: failed receiving information from client")
			removeFile("videos/", filename)
			return
		}
	}
}

// Send json response after file upload
func writeJsonResponse(w http.ResponseWriter, filename string) (vd.Vidinfo, error) {
	var (
		data    vd.Data
		vidinfo vd.Vidinfo
		err     error
		info    []byte
	)

	vidinfo, err = tc.GetVidInfo("./videos/", filename, CONF.TempJson, CONF.DataGen, CONF.TempTxt)
	if err != nil {
		log.Println(err)
		return vidinfo, err
	}

	if CONF.Presets {
		data, err = db.AddPresetsToJson(vidinfo)
		if err != nil {
			return vidinfo, err
		}

		info, err = json.Marshal(data)
		if err != nil {
			log.Println(err)
			return vidinfo, err
		}
	} else {
		info, err = json.Marshal(vidinfo)
		if err != nil {
			log.Println(err)
			return vidinfo, err
		}
	}

	w.WriteHeader(200)
	w.Write(info)

	return vidinfo, nil
}

func tctypeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(444)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	type response struct {
		Typechange bool `json:"tc"`
	}
	var rsp response
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&rsp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if rsp.Typechange {
		CONF.Presets = false
		uploadtemplate = "./views/templates/uploadcl.html"
	} else {
		CONF.Presets = true
		uploadtemplate = "./views/templates/upload.html"
	}

	w.WriteHeader(200)
}

func transcodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(444)
		return
	}

	if err := r.ParseForm(); err != nil {
		crGot = 2
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	// Decode json file
	if CONF.Presets {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&prd)
		if err != nil {
			crGot = 2
			log.Println(err)
			w.WriteHeader(500)
			return
		}
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&vf)
		if err != nil {
			crGot = 2
			log.Println(err)
			w.WriteHeader(500)
			return
		}
	}

	w.WriteHeader(200)
	crGot = 1
}

func removeFile(path string, filename string) {
	resetData()
	if os.Remove(path+filename) != nil {
		lp.WLog("Error: failed removing source file")
	}
	db.RemoveColumnByName(filename, "Video")
	return
}

func resetData() {
	crGot = 0
	vf = vd.Video{}
	prd = vd.PData{}
}

func main() {

	// Make a new Broker instance
	lp.B = &lp.Broker{
		Clients:        make(map[chan string]bool),
		NewClients:     make(chan (chan string)),
		DefunctClients: make(chan (chan string)),
		Messages:       make(chan string),
	}

	// Start processing events
	lp.B.Start()

	// Load config file
	var err error
	CONF, err = cf.GetConf()
	if err != nil {
		log.Println("Error: failed to load config file")
		log.Println(err)
		return
	}

	//Use client choices html if CONF.Presets false
	if CONF.Advanced && !CONF.Presets {
		CONF.Presets = false
		uploadtemplate = "./views/templates/uploadcl.html"
	} else {
		uploadtemplate = "./views/templates/upload.html"
	}

	// Write all logs to file
	err = lp.OpenLogFile(CONF.LogP)
	if err != nil {
		log.Println("Error: failed open log file")
		log.Panicln(err)
		return
	}
	defer lp.LogFile.Close()

	//Open database
	err = db.OpenDatabase()
	if err != nil {
		//log.Panicln(err)
		return
	}

	// Insert CONF.Presets to database
	err = db.InsertPresets()
	if err != nil {
		log.Println("Error: failed to insert CONF.Presets to database")
		log.Panicln(err)
		return
	}

	http.Handle("/transcode", http.HandlerFunc(transcodeHandler))
	http.Handle("/tctype", http.HandlerFunc(tctypeHandler))
	http.Handle("/sse/dashboard", lp.B)
	http.Handle("/upload", http.HandlerFunc(uploadHandler))
	http.Handle("/", http.FileServer(http.Dir("views")))
	fmt.Println("Listening on port: 8080...")
	log.Fatalf("Exited: %s", http.ListenAndServe(":8080", nil))
}
