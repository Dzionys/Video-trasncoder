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

	cf "../conf"
	db "../database"
	"../lp"
	vd "../videodata"
)

var (
	wg      sync.WaitGroup
	DEBUG   = false
	CONF    cf.Config
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

func getRatio(res string, duration int, clid string) {
	i := strings.Index(res, "time=")
	if i >= 0 {
		time := res[i+5:]
		if len(time) > 8 {
			time = time[0:8]
			sec := durToSec(time)
			per := (sec * 100) / duration
			if lastPer != per {
				lastPer = per
				lp.UpdateLogMessage(fmt.Sprintf("Progress: %v %%", per), clid)
			}
			allRes = ""
		}
	}
}

func runCmdCommand(cmdl string, dur string, wg *sync.WaitGroup, clid string) error {
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
		lp.WLog("Progress bar unavailable", clid)
	} else {
		duration := durToSec(dur)
		for {
			_, err := stdout.Read(oneByte)
			if err != nil {
				log.Println(err)
				break
			}
			allRes += string(oneByte)
			getRatio(allRes, duration, clid)
		}
	}

	return nil
}

func ProcessVodFile(source string, data vd.Vidinfo, cldata vd.Video, prdata vd.PData, conf cf.Config, clid string) {
	lp.WLog("Starting VOD Processor..", clid)
	var (
		err error
		cmd string
	)

	CONF = conf

	// Path to source file
	sfpath := CONF.SD + source

	// Checks if source file exists
	if source != "" {
		if _, err := os.Stat(sfpath); err == nil {
			lp.WLog("File found", clid)
		} else if os.IsNotExist(err) {
			lp.WLog("Error: file does not exist", clid)
			return
		} else {
			log.Println(err)
			lp.WLog("Error: file may or may not exist", clid)
			removeFile("videos/", source, clid)
			return
		}
	} else {
		removeFile("videos/", source, clid)
		return
	}

	// Full source file name
	fullsfname, err := filepath.EvalSymlinks(sfpath)
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to get full file name", clid)
		removeFile("videos/", source, clid)
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
		lp.WLog(fmt.Sprintf("Error: file \"%v\" already transcoding", sfnamewe+".mp4"), clid)
		removeFile("videos/", source, clid)
		return
	} else if _, err := os.Stat(destinationfile); err == nil {
		lp.WLog(fmt.Sprintf("Error: file \"%v\" already exist in transcoded folder", sfnamewe+".mp4"), clid)
		removeFile("videos/", source, clid)
		return
	}

	lp.WLog(fmt.Sprintf("Starting to process %s", source), clid)

	// If data is empty get video info
	if data.IsEmpty() {
		data, err = GetVidInfo(CONF.SD, source, CONF.TempJson, CONF.DataGen, CONF.TempTxt, clid)
		if err != nil {
			log.Println(err)
			removeFile("videos/", source, clid)
			return
		}
	}

	msg := "%v video track(s), %v audio track(s) and %v subtitle(s) found"
	frmt := fmt.Sprintf(msg, data.Videotracks, data.Audiotracks, data.Subtitles)
	lp.WLog(frmt, clid)

	// Generate command line
	var save bool
	var tempdfs []string
	if CONF.Advanced {
		if CONF.Presets {
			save = prdata.Save
			cmd, tempdfs, err = generatePresetCmdLine(prdata, data, sfpath, fullsfname, fmt.Sprintf("%v%v", CONF.TD, sfnamewe))
			tempfile = tempdfs[0]
			if err != nil {
				lp.WLog("Error: failed to generate cmd line", clid)
				log.Println(err)
				removeFile("videos/", source, clid)
				return
			}
		} else {
			save = cldata.Save
			cmd = generateClientCmdLine(cldata, data, sfpath, fullsfname, tempfile)
		}
	} else {
		cmd = generateBaseCmdLine(data, sfpath, tempfile, fullsfname)
	}

	// check if client wants to save cmd line
	if save {
		err := db.AddCmdLine(source, cmd, tempdfs)
		if err != nil {
			lp.WLog("Error: failed to insert command line in database", clid)
			log.Println(err)
			removeFile(CONF.SD, source, clid)
		} else {
			lp.WLog("Transcoding parameters saved", clid)
		}
	} else {
		var dfsl string
		for i, d := range tempdfs {
			if i != len(tempdfs)-1 {
				dfsl += d + " "
			} else {
				dfsl += d
			}
		}
		go StartTranscode(source, CONF, cmd, dfsl, clid)
	}
}

func StartTranscode(source string, conf cf.Config, cmdg string, dfsl string, clid string) {
	var (
		err     error
		cmd     string
		dfsline string
		dfs     []string
		dur     string
		data    vd.Vidinfo
	)

	CONF = conf

	// Path to source file
	sfpath := CONF.SD + source

	// Checks if source file exists
	if source != "" {
		if _, err := os.Stat(sfpath); err == nil {
			lp.WLog("File found", clid)
		} else if os.IsNotExist(err) {
			lp.WLog("Error: file does not exist", clid)
			return
		} else {
			log.Println(err)
			lp.WLog("Error: file may or may not exist", clid)
			removeFile("videos/", source, clid)
			return
		}
	} else {
		removeFile("videos/", source, clid)
		return
	}

	// Full source file name
	fullsfname, err := filepath.EvalSymlinks(sfpath)
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to get full file name", clid)
		removeFile("videos/", source, clid)
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

	data, err = GetVidInfo("./videos/", source, CONF.TempJson, CONF.DataGen, CONF.TempTxt, clid)

	// ===============================================================

	if cmdg != "" {
		cmd = cmdg
		dfsline = dfsl
	} else {
		cmd, dfsline, err = db.GetTranscodingInfo(source)
	}

	tempdfs := strings.Split(dfsline, " ")

	if dfsline != "" {
		tempfile = tempdfs[0]

		// removes path from stream files names
		for _, d := range tempdfs {
			df := strings.SplitAfterN(d, "/", 3)[2]
			dfs = append(dfs, df)
		}
	}

	// Run generated command line
	lp.WLog("Starting to transcode", clid)
	if CONF.Debug {
		dur = CONF.DebugEnd
	} else {
		dur = data.Videotrack[0].Duration
	}

	err = db.UpdateState(source, "Transcoding")
	if err != nil {
		lp.WLog("Error: failed to update state in database", clid)
		log.Println(err)
		removeFile(CONF.SD, source, clid)
		return
	}

	wg.Add(1)
	err = runCmdCommand(cmd, dur, &wg, clid)
	wg.Wait()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: could not start trancoding", clid)
		log.Printf("Error cmd line: %v", cmd)
		removeFile("videos/", source, clid)
		return
	} else if out, err := os.Stat(tempfile); os.IsNotExist(err) || out == nil {
		log.Println(err)
		lp.WLog("Error: transcoder failed", clid)
		log.Printf("Error cmd line: %v", cmd)
		removeFile("videos/", source, clid)
		return
	} else {

		if _, err = os.Stat(CONF.DD); os.IsNotExist(err) {
			os.Mkdir(CONF.DD, 0777)
		}
		// Removes source file and moves transcoded file to /videos/transcoded
		if CONF.Advanced && CONF.Presets {
			var (
				ndata []vd.Vidinfo
			)
			removeFile(CONF.SD, source, clid)
			for i, _ := range tempdfs {
				os.Rename(CONF.TD+dfs[i], CONF.DD+dfs[i])
				nd, err := GetVidInfo(CONF.DD, dfs[i], CONF.TempJson, CONF.DataGen, CONF.TempTxt, clid)
				if err != nil {
					lp.WLog("Error: failed getting video data", clid)
					log.Println(err)
					removeStreamFiles(CONF.DD, dfs, sfnamewe, clid)
					return
				}
				ndata = append(ndata, nd)
			}
			err = db.InsertStream(ndata, dfs, "Transcoded", sfnamewe)
			if err != nil {
				lp.WLog("Error: failed to insert stream data in database", clid)
				log.Println(err)
				removeStreamFiles(CONF.DD, dfs, sfnamewe, clid)
				return
			}

			msg := fmt.Sprintf("Transcoding coplete, stream name: %v", sfnamewe)
			lp.WLog(msg, clid)

		} else {
			removeFile(CONF.SD, source, clid)
			os.Rename(tempfile, destinationfile)
			dfn := sfnamewe + ".mp4"
			ndata, err := GetVidInfo(CONF.DD, dfn, CONF.TempJson, CONF.DataGen, CONF.TempTxt, clid)
			if err != nil {
				lp.WLog("Error: failed getting video data", clid)
				log.Println(err)
				removeFile(CONF.DD, dfn, clid)
				return
			}
			err = db.InsertVideo(ndata, dfn, "Transcoded", -1)
			if err != nil {
				lp.WLog("Error: failed to insert video data in database", clid)
				log.Println(err)
				removeFile(CONF.DD, dfn, clid)
				return
			}

			msg := fmt.Sprintf("Transcoding coplete, file name: %v", filepath.Base(tempfile))
			lp.WLog(msg, clid)
		}
	}

}

func removeFile(path string, filename string, clid string) {
	if _, err := os.Stat(path + filename); os.Remove(path+filename) != nil && !os.IsNotExist(err) {
		lp.WLog("Error: failed removing file", clid)
	}
	db.RemoveRowByName(filename, "Video")
	return
}

func removeStreamFiles(path string, filenames []string, sname string, clid string) {
	for _, fn := range filenames {
		if os.Remove(path+fn) != nil {
			lp.WLog("Error: failed removing stream file(s)", clid)
		}
	}
	db.RemoveRowByName(sname, "Stream")
	return
}
