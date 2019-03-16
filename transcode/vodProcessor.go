package transcoder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"../lp"
	"github.com/BurntSushi/toml"
	logrus "github.com/Sirupsen/logrus"
)

var (
	wg    sync.WaitGroup
	DEBUG = false
	CONF  Config
)

func generateCmdLine(d Vidinfo, sf string, df string, sfname string) string {
	var (
		cmd     string
		mapping string
		vcode   string
		acode   string
		scode   string
		frate   string
		ss      string
	)

	// Video cmd
	if d.videotrack[0].frameRate < 25 {
		baseln := "-r 25 -filter_complex [0:v]setpts=%v/25*PTS[v];[0:a]atempo=25/%v[a] -map [v] -map [a]"
		frate = fmt.Sprintf(baseln, d.videotrack[0].frameRate, d.videotrack[0].frameRate)
	} else {
		frate = ""
		mapping = fmt.Sprintf("-map 0:%v", d.videotrack[0].index)
	}
	if d.videotrack[0].width > 1280 || d.videotrack[0].codecName != "h265" {
		vcode = "-c:v:0 libx265 -x265-params \"preset=slower:me=hex:no-rect=1:no-amp=1:rd=4:aq-mode=2:"
		vcode += "aq-strength=0.5:psy-rd=1.0:psy-rdoq=0.2:bframes=3:min-keyint=1\" "
		vcode += fmt.Sprintf("-b:v:0 %vk -metadata:s:v:0 name=\"%v\"", CONF.VBW, sfname)
		if d.videotrack[0].width > 1280 {
			vcode += " -filter:v:0 \"scale=iw*sar:ih,pad=iw:iw/16*9:0:(oh-ih)/2\" -aspect 16:9"
		}
	} else {
		vcode = fmt.Sprintf(" -c:v:0 copy -metadata:s:v:0 name=\"%v\"", sfname)
	}

	// Audio cmd
	acode = ""
	for i := 0; i < d.audiotracks; i++ {
		if !(d.videotrack[0].frameRate < 25) {
			mapping += fmt.Sprintf(" -map 0:%v", d.audiotrack[i].index)
		}
		acode += fmt.Sprintf(" -c:a:%v libfdk_aac -ac 2 -b:a:%v %vk -metadata language=%v", i, i, CONF.ABW, d.audiotrack[i].language)
	}

	// Subtitles cmd
	scode = ""
	if d.subtitles > 0 {
		for i := 0; i < d.subtitles; i++ {
			scode += fmt.Sprintf(" -c:s:%v copy -metadata:s:s:%v language=%v", d.subtitle[i].index, d.subtitle[i].index, d.subtitle[i].language)
		}
	}

	// Set debug duration if debug is selected
	ss = ""
	if DEBUG {
		ss = CONF.DebugStart + " " + CONF.DebugEnd
	}

	// Add all parts in one command line
	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v %v -async 1 -vsync 1 %v", sf, ss, mapping, frate, vcode, acode, scode, df)
	lp.WLog("Cmd line generated")

	return cmd
}

func runCmdCommand(cmd string, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Splits cmd command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	cmdl := exec.Command(head, parts...)
	cmdl.Stdout = logrus.StandardLogger().Out
	cmdl.Stderr = logrus.StandardLogger().Out
	if err := cmdl.Run(); err != nil {
		return err
	}

	return nil
}

func ProcessVodFile(source string, debug bool) {
	lp.WLog("Starting VOD Processor..")
	var err error

	// Load config file
	CONF, err = upConf()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to load config file")
		return
	}

	DEBUG = debug
	sfpath := CONF.SD + source

	// Checks if source file exists
	if source != "" {
		if _, err := os.Stat(sfpath); err == nil {
			lp.WLog("File found")
		} else if os.IsNotExist(err) {
			lp.WLog("Error: file does not exist")
			return
		} else {
			log.Println(err)
			lp.WLog("Error: file may or may not exist")
			return
		}
	} else {
		return
	}

	// Full source file name
	fullsfname, err := filepath.EvalSymlinks(sfpath)
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to get full file name")
		return
	}

	// Source file extension
	sfext := filepath.Ext(fullsfname)

	// Source file name without extension
	sfnamewe := strings.Split(source, sfext)[0]

	// File name after transcoding
	if _, err = os.Stat(CONF.DD); os.IsNotExist(err) {
		os.Mkdir(CONF.DD, 0777)
	}
	destinationfile := fmt.Sprintf("%v%v.mp4", CONF.DD, sfnamewe)

	// Checks if transcoded file with the same name already exists
	if _, err := os.Stat(destinationfile); err == nil {
		lp.WLog(fmt.Sprintf("Error: file \"%v\" already exists", fullsfname))
		return
	}

	lp.WLog(fmt.Sprintf("Starting to process %s", source))

	// Geting data about video
	wg.Add(1)
	infob, err := GetMediaInfoJson(sfpath, &wg)
	if err != nil {
		log.Println(err)
		lp.WLog(fmt.Sprintf("Error: could not get json data from file - %v", fullsfname))
		return
	}
	wg.Wait()

	// Writing data to temporary json file
	var raw map[string]interface{}
	json.Unmarshal(infob, &raw)
	info, _ := json.Marshal(raw)
	err = ioutil.WriteFile(CONF.TempJson, info, 0666)
	if err != nil {
		fmt.Println(err)
		lp.WLog("Error: could not create json file")
		return
	}

	// Run python file to get nesessary data from json file
	gpath, err := filepath.Abs(CONF.DataGen)
	wg.Add(1)
	err = generateDataFile(&wg, gpath)
	wg.Wait()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to generate video data")
		return
	}

	// Write data to Vidinfo struct
	data, err := ParseFile(CONF.TempTxt)
	if err != nil {
		lp.WLog("Error: failed parsing data file")
		return
	}

	msg := "%v video track(s), %v audio track(s) and %v subtitle(s) found"
	frmt := fmt.Sprintf(msg, data.videotracks, data.audiotracks, data.subtitles)
	lp.WLog(frmt)

	// Generate command line
	cmd := []byte(generateCmdLine(data, sfpath, destinationfile, fullsfname))

	// Run generated command line
	lp.WLog("Starting to transcode")
	wg.Add(1)
	err = runCmdCommand(string(cmd), &wg)
	if err != nil {
		log.Println(err)
		lp.WLog("Error: could not start trancoding")
		return
	}
	wg.Wait()

	msg = fmt.Sprintf("Transcoding coplete, file name: %v", filepath.Base(destinationfile))
	lp.WLog(msg)
}

// Load config file
func upConf() (Config, error) {
	var conf Config
	if _, err := toml.DecodeFile("transcode/conf.toml", &conf); err != nil {
		log.Println("error geting conf.toml: ", err)
		return conf, err
	}
	return conf, nil
}
