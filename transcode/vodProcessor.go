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
)

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
				sse.UpdateLogMessage(fmt.Sprintf("Progress: %v %%", per))
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

	// Run commad
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return err
	}
	oneByte := make([]byte, 8)

	// If duration is not provided dont sent progress bar
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

func ProcessVodFile(source string, data Vidinfo, cldata Video) {
	lp.WLog("Starting VOD Processor..")
	var (
		err error
		dur string
		cmd string
	)

	// Path to source file
	sfpath := CONF.SD + source

	// Load config file
	CONF, err = upConf()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to load config file")
		removeFile(sfpath)
		return
	}

	// Checks if source file exists
	if source != "" {
		if _, err := os.Stat(sfpath); err == nil {
			lp.WLog("File found")
		} else if os.IsNotExist(err) {
			lp.WLog("Error: file does not exist")
			removeFile(sfpath)
			return
		} else {
			log.Println(err)
			lp.WLog("Error: file may or may not exist")
			removeFile(sfpath)
			return
		}
	} else {
		removeFile(sfpath)
		return
	}

	// Full source file name
	fullsfname, err := filepath.EvalSymlinks(sfpath)
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to get full file name")
		removeFile(sfpath)
		return
	}

	// Source file name without extension
	sfnamewe := strings.Split(source, filepath.Ext(fullsfname))[0]

	// If transcoding directory does not exist creat it
	if _, err = os.Stat(CONF.TD); os.IsNotExist(err) {
		os.Mkdir(CONF.TD, 0777)
	}

	// File name after transcoding
	tempfile := fmt.Sprintf("%v%v.mp4", CONF.TD, sfnamewe)

	// f
	destinationfile := fmt.Sprintf("%v%v.mp4", CONF.DD, sfnamewe)

	// Checks if transcoded file with the same name already exists
	if _, err := os.Stat(tempfile); err == nil {
		lp.WLog(fmt.Sprintf("Error: file \"%v\" already transcoding", sfnamewe+".mp4"))
		removeFile(sfpath)
		return
	} else if _, err := os.Stat(destinationfile); err == nil {
		lp.WLog(fmt.Sprintf("Error: file \"%v\" already exist in transcoded folder", sfnamewe+".mp4"))
		removeFile(sfpath)
		return
	}

	lp.WLog(fmt.Sprintf("Starting to process %s", source))

	// If data is empty get video info
	if data.IsEmpty() {
		data, err = GetVidInfo(sfpath, CONF.TempJson, CONF.DataGen, CONF.TempTxt)
		if err != nil {
			log.Println(err)
			removeFile(sfpath)
			return
		}
	}

	msg := "%v video track(s), %v audio track(s) and %v subtitle(s) found"
	frmt := fmt.Sprintf(msg, data.Videotracks, data.Audiotracks, data.Subtitles)
	lp.WLog(frmt)

	// Generate command line
	if CONF.Advanced {
		cmd, _ = generateClientCmdLine(cldata, data, sfpath, tempfile, fullsfname)
	} else {
		cmd = generateBaseCmdLine(data, sfpath, tempfile, fullsfname)
	}

	// Run generated command line
	lp.WLog("Starting to transcode")
	if CONF.Debug {
		dur = CONF.DebugEnd
	} else {
		dur = data.Videotrack[0].Duration
	}
	wg.Add(1)
	err = runCmdCommand(cmd, dur, &wg)
	wg.Wait()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: could not start trancoding")
		removeFile(sfpath)
		return
	} else if out, err := os.Stat(tempfile); os.IsNotExist(err) || out == nil {
		log.Println(err)
		lp.WLog("Error: transcoder failed")
		removeFile(sfpath)
		return
	} else {

		// Removes source file and moves transcoded file to /videos/transcoded
		os.Remove(sfpath)
		if _, err = os.Stat(CONF.DD); os.IsNotExist(err) {
			os.Mkdir(CONF.DD, 0777)
		}
		os.Rename(tempfile, destinationfile)
		os.Remove(tempfile)

		msg = fmt.Sprintf("Transcoding coplete, file name: %v", filepath.Base(tempfile))
		lp.WLog(msg)
	}
}

func removeFile(filepath string) {
	if os.Remove(filepath) != nil {
		lp.WLog("Error: failed to remove source file")
	}
	return
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
