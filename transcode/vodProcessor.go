package transcoder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"../lp"
	"../sse"
	"github.com/BurntSushi/toml"
)

var (
	wg      sync.WaitGroup
	DEBUG   = false
	CONF    Config
	allRes  = ""
	lastPer = -1
	//boiz    []Vid
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
	if d.Videotrack[0].FrameRate < 25 {
		baseln := "-r 25 -filter_complex [0:v]setpts=%v/25*PTS[v];[0:a]atempo=25/%v[a] -map [v] -map [a]"
		frate = fmt.Sprintf(baseln, d.Videotrack[0].FrameRate, d.Videotrack[0].FrameRate)
	} else {
		frate = ""
		mapping = fmt.Sprintf("-map 0:%v", d.Videotrack[0].Index)
	}
	if d.Videotrack[0].Width > 1280 || d.Videotrack[0].CodecName != "h265" {
		vcode = "-c:v:0 libx265 -x265-params \"preset=slower:me=hex:no-rect=1:no-amp=1:rd=4:aq-mode=2:"
		vcode += "aq-strength=0.5:psy-rd=1.0:psy-rdoq=0.2:bframes=3:min-keyint=1\" "
		vcode += fmt.Sprintf("-b:v:0 %vk -metadata:s:v:0 name=\"%v\"", CONF.VBW, sfname)
		if d.Videotrack[0].Width > 1280 {
			vcode += " -filter:v:0 \"scale=iw*sar:ih,pad=iw:iw/16*9:0:(oh-ih)/2\" -aspect 16:9"
		}
	} else {
		vcode = fmt.Sprintf(" -c:v:0 copy -metadata:s:v:0 name=\"%v\"", sfname)
	}

	// Audio cmd
	// acode = ""
	// for i := 0; i < d.Sudiotracks; i++ {
	// 	if !(d.videotrack[0].frameRate < 25) {
	// 		mapping += fmt.Sprintf(" -map 0:%v", d.Audiotrack[i].Index)
	// 	}
	// 	acode += fmt.Sprintf(" -c:a:%v libfdk_aac -ac 2 -b:a:%v %vk -metadata language=%v", i, i, CONF.ABW, d.Audiotrack[i].Language)
	// }

	// Subtitles cmd
	scode = ""
	if d.Subtitles > 0 {
		for i := 0; i < d.Subtitles; i++ {
			scode += fmt.Sprintf(" -c:s:%v copy -metadata:s:s:%v language=%v", d.Subtitle[i].Index, d.Subtitle[i].Index, d.Subtitle[i].Language)
		}
	}

	// Set debug duration if debug is selected
	ss = ""
	if DEBUG {
		ss = "-ss " + CONF.DebugStart + " -t " + CONF.DebugEnd
	}

	// Add all parts in one command line
	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v %v -async 1 -vsync 1 %v", sf, ss, mapping, frate, vcode, acode, scode, df)
	lp.WLog("Cmd line generated")

	return cmd
}

func durToSec(dur string) (sec int) {
	durAry := strings.Split(dur, ":")
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.Atoi(durAry[0])
	sec = hr * (60 * 60)
	min, _ := strconv.Atoi(durAry[1])
	sec += min * (60)
	second, _ := strconv.Atoi(durAry[2])
	sec += second
	return
}
func getRatio(res string, duration int) {
	i := strings.Index(res, "time=")
	if i >= 0 {
		time := res[i+5:]
		if len(time) > 8 {
			time = time[0:8]
			sec := durToSec(time)
			per := (sec * 100) / duration
			if lastPer != per {
				lastPer = per
				sse.UpdateLogMessage(fmt.Sprintf("Percentage done: %v proc.", per))
			}

			allRes = ""
		}
	}
}

func runCmdCommand(cmdl string, dur string, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Splits cmd command
	parts := strings.Fields(cmdl)
	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)

	// Creates pipe to listen to output
	stdout, err := cmd.StderrPipe()
	if err != nil {
		log.Println(err)
		return err
	}

	if err := cmd.Start(); err != nil {
		log.Println(err)
		return err
	}
	oneByte := make([]byte, 8)

	if dur == "" {
		lp.WLog("Progress bar unavailable")
	} else {
		duration := durToSec(dur)
		for {
			_, err := stdout.Read(oneByte)
			if err != nil {
				log.Println(err)
				break
			}
			allRes += string(oneByte)
			getRatio(allRes, duration)
		}
	}

	return nil
}

func ProcessVodFile(source string, debug bool) {
	lp.WLog("Starting VOD Processor..")
	var (
		err error
		dur string
	)

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

	// Get video info
	data, err := GetVidInfo(sfpath, CONF.TempJson, CONF.DataGen, CONF.TempTxt)
	if err != nil {
		log.Println(err)
		return
	}

	msg := "%v video track(s), %v audio track(s) and %v subtitle(s) found"
	frmt := fmt.Sprintf(msg, data.Videotracks, data.Audiotracks, data.Subtitles)
	lp.WLog(frmt)

	// Generate command line
	cmd := []byte(generateCmdLine(data, sfpath, destinationfile, fullsfname))

	// Run generated command line
	lp.WLog("Starting to transcode")
	wg.Add(1)
	if debug {
		dur = CONF.DebugEnd
	} else {
		dur = data.Videotrack[0].Duration
	}
	err = runCmdCommand(string(cmd), dur, &wg)
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
