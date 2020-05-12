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
	"strings"
	"sync"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"

	cf "./conf"
	"./controllers"
	db "./database"
	"./database/auth"
	"./lp"
	tc "./transcode"
	vd "./videodata"
)

var (
	uploadtemplate string
	tcvidpath      = "/home/dzionys/Documents/Video-trasncoder/videos/transcoded/%v"
	basetemplate   = "./views/templates/base.html"
	wg             sync.WaitGroup
	// CONF contains configuration info
	CONF   cf.Config
	vfnprd = make(chan vd.VfNPrd)
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		t, err := template.ParseFiles(basetemplate, uploadtemplate)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
		}

	case "POST":

		lp.WLog("Upload started", r.RemoteAddr)

		//Starts readig file by chuncking
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			lp.WLog("Error: failed to upload file", r.RemoteAddr)
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
			lp.WLog("Error: this file format is not allowed "+filepath.Ext(handler.Filename), r.RemoteAddr)
			w.WriteHeader(403)
			return
		}

		// Checks if uploaded file with the same name already exists
		if _, err := os.Stat("./videos/" + handler.Filename); err == nil {
			lp.WLog(fmt.Sprintf("Error: file \"%v\" already exists", handler.Filename), r.RemoteAddr)
			w.WriteHeader(403)
			return
		}

		//Create empty file in /videos folder
		lp.WLog("Creating file", r.RemoteAddr)
		dst, err := os.OpenFile("./videos/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		defer dst.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			lp.WLog("Error: could not create file", r.RemoteAddr)
			return
		}

		//Copies a temporary file to empty file in /videos folder
		lp.WLog("Writing to file", r.RemoteAddr)
		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			lp.WLog("Error: failed to write file", r.RemoteAddr)
			removeFile("./videos/", handler.Filename, r.RemoteAddr)
			return
		}
		lp.WLog("Upload successful", r.RemoteAddr)

		data, err := writeJSONResponse(w, handler.Filename, r.RemoteAddr)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			lp.WLog("Error: failed send video data to client", r.RemoteAddr)
			removeFile("./videos/", handler.Filename, r.RemoteAddr)
			return
		}

		err = db.InsertVideo(data, handler.Filename, "Not transcoded", -1)
		if err != nil {
			lp.WLog("Error: failed to insert video data in database", r.RemoteAddr)
			log.Println(err)
			w.WriteHeader(500)
			removeFile("./videos/", handler.Filename, r.RemoteAddr)
			return
		}

		lp.UpdateMessage(handler.Filename)

		if CONF.Advanced {
			go func() {
				dat := <-vfnprd
				if dat.Err == nil {
					vf := dat.Video
					prd := dat.PData
					go tc.ProcessVodFile(handler.Filename, data, vf, prd, CONF, r.RemoteAddr)
				}
			}()
			resetData()
		} else {
			//go tc.ProcessVodFile(handler.Filename, data, vf, prd, CONF, r.RemoteAddr)
			resetData()
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Send json response after file upload
func writeJSONResponse(w http.ResponseWriter, filename string, clid string) (vd.Vidinfo, error) {
	var (
		data    vd.Data
		vidinfo vd.Vidinfo
		err     error
		info    []byte
	)

	vidinfo, err = tc.GetVidInfo("./videos/", filename, CONF.TempJson, CONF.DataGen, CONF.TempTxt, clid)
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
		w.WriteHeader(400)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(422)
		return
	}

	type response struct {
		Typechange string `json:"Tc"`
	}
	var rsp response
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&rsp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(415)
		return
	}

	if rsp.Typechange == "true" {
		CONF.Presets = false
		uploadtemplate = "./views/templates/uploadcl.html"
	} else if rsp.Typechange == "false" {
		CONF.Presets = true
		uploadtemplate = "./views/templates/upload.html"
	} else {
		log.Println(fmt.Errorf("uknown change type: '%v', expected 'true' or 'false'", rsp.Typechange))
		w.WriteHeader(415)

	}

	w.WriteHeader(200)
}

func transcodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
		return
	}

	var err error

	err = r.ParseForm()
	if err != nil {
		log.Println(err)
		w.WriteHeader(422)
		return
	}

	var (
		vf  vd.Video
		prd vd.PData
	)

	// Decode json file
	if CONF.Presets {
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&prd)
		if err != nil {
			log.Println(err)
			w.WriteHeader(415)
			return
		}
	} else {
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&vf)
		if err != nil {
			log.Println(err)
			w.WriteHeader(415)
			return
		}
	}

	data := vd.VfNPrd{
		prd,
		vf,
		err,
	}

	vfnprd <- data

	w.WriteHeader(200)
}

func vdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(400)
		return
	}

	data, err := db.PutVideosToJson()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	dt, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dt)
}

func ngxMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(400)
		return
	}
	var sqncs vd.Sequences

	vars := mux.Vars(r)

	if filepath.Ext(vars["name"]) == ".mp4" {
		err := db.IsExist("Video", vars["name"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}

		temp := vd.Clip{
			"source",
			fmt.Sprintf(tcvidpath, vars["name"]),
		}
		var tempclip vd.Clips
		tempclip.Clips = append(tempclip.Clips, temp)
		sqncs.Sequences = append(sqncs.Sequences, tempclip)

	} else {
		names, err := db.GetAllStreamVideos(vars["name"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}

		for _, n := range names {
			temp := vd.Clip{
				"source",
				fmt.Sprintf(tcvidpath, n),
			}
			var tempclip vd.Clips
			tempclip.Clips = append(tempclip.Clips, temp)
			sqncs.Sequences = append(sqncs.Sequences, tempclip)
		}
	}

	j, err := json.Marshal(sqncs)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write(j)
}

func vidUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
		return
	}

	var (
		updata vd.Update
		err    error
	)

	if err = r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(422)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&updata)
	if err != nil {
		log.Println(err)
		w.WriteHeader(415)
		w.Write([]byte("error: failed to decode json file"))
		return
	}

	if updata.Utype == 2 {
		newName := strings.Split(updata.Data, "/")[0]
		oldName := strings.Split(updata.Data, "/")[1]

		err = db.UpdateVideoName(newName, oldName, updata.Stream)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			w.Write([]byte("error: failed to update video name"))
			return
		} else {
			w.Write([]byte("video name updated successfully"))
		}
	} else if updata.Utype == 1 {
		err = db.RemoveVideo(updata.Data, updata.Stream)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			w.Write([]byte("error: failed to remove video(s)"))
			return
		} else {
			w.Write([]byte("video deleted successfully"))
		}
	} else if updata.Utype == 3 {
		if updata.Data == "" {
			w.WriteHeader(422)
			w.Write([]byte("error: video name has not been provided"))
			return
		} else if !FileExist(updata.Data) {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("error: video: %v not found", updata.Data)))
			return
		}
		go tc.StartTranscode(updata.Data, CONF, "", "", r.RemoteAddr)
	} else if updata.Utype == 0 {
		w.WriteHeader(415)
		w.Write([]byte("error: JSON file is not valid or wrong format"))
		return
	} else {
		w.WriteHeader(422)
		w.Write([]byte("error: unsupported update number " + string(updata.Utype)))
		return
	}
}

// FileExist checks if file exist in database
func FileExist(name string) bool {
	notstream := false
	for _, ave := range CONF.FileTypes {
		if filepath.Ext(name) == ave {
			notstream = true
		}
	}

	var exist error
	if notstream {
		exist = db.IsExist("Video", name)
	} else {
		exist = db.IsExist("Stream", name)
	}

	if exist == nil {
		return true
	}
	return false
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
	watchtemplate := template.Must(template.ParseGlob("./views/templates/watch.html"))
	err := watchtemplate.Execute(w, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	listtemplate := template.Must(template.ParseGlob("./views/templates/list.html"))
	err := listtemplate.Execute(w, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
}

func removeFile(path string, filename string, clid string) {
	resetData()
	if os.Remove(path+filename) != nil {
		lp.WLog("Error: failed removing source file", clid)
	}
	db.RemoveRowByName(filename, "Video")
	return
}

func resetData() {
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		next.ServeHTTP(w, r)
	})
}

func main() {

	// Make a new Broker instance
	lp.B = &lp.Broker{
		Clients:        make(map[chan string]string),
		NewClients:     make(chan lp.Client),
		DefunctClients: make(chan (chan string)),
		Messages:       make(chan lp.Message),
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

	err = godotenv.Load()
	if err != nil {
		log.Println("Error: failed to load .env file")
		log.Println(err)
		return
	}
	port := os.Getenv("PORT")

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
		log.Panicln(err)
		return
	}

	// Insert CONF.Presets to database
	err = db.InsertPresets()
	if err != nil {
		log.Println("Error: failed to insert CONF.Presets to database")
		log.Panicln(err)
		return
	}

	// Handle routes
	//http.Handle("/", routes.Handlers())

	r := mux.NewRouter().StrictSlash(true)
	//r.Use(commonMiddleware)

	r.Handle("/ngx/mapping/{name}", http.HandlerFunc(ngxMappingHandler))
	r.Handle("/transcode", http.HandlerFunc(transcodeHandler))
	r.Handle("/tctype", http.HandlerFunc(tctypeHandler))
	r.Handle("/vd", http.HandlerFunc(vdHandler))
	r.Handle("/list", http.HandlerFunc(listHandler))
	r.Handle("/videoupdate", http.HandlerFunc(vidUpdateHandler))
	r.Handle("/watch", http.HandlerFunc(playerHandler))
	r.Handle("/sse/dashboard", lp.B)
	r.Handle("/upload", http.HandlerFunc(uploadHandler))
	r.HandleFunc("/register", controllers.CreateUser)
	r.HandleFunc("/login", controllers.Login)
	// Auth route
	s := r.PathPrefix("/auth").Subrouter()
	s.Use(auth.JwtVerify)
	//s.HandleFunc("/user", controllers.FetchUsers).Methods("GET")
	s.HandleFunc("/user/{id}", controllers.GetUser).Methods("GET")
	s.HandleFunc("/user/{id}", controllers.UpdateUser).Methods("PUT")
	s.HandleFunc("/user/{id}", controllers.DeleteUser).Methods("DELETE")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("views"))))

	fmt.Printf("Listening on port: %v...\n", port)
	log.Fatalf("Exited: %s", http.ListenAndServe(":"+port, r))
}
