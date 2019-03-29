package transcoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"../lp"
	vd "../videodata"
)

func GetMediaInfoJson(source string, wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()

	cmd := "ffprobe -v quiet -print_format json -show_streams"
	cmd += " " + source

	//Splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		return out, err
	}

	if fmt.Sprintf("%s", out) == "" {
		return out, errors.New("json data is empty")
	}

	//checks if data is json file
	if !json.Valid([]byte(out)) {
		return out, errors.New("data is not valid json file")
	}

	return out, nil
}

func generateDataFile(wg *sync.WaitGroup, gpath string) error {
	defer wg.Done()

	out, err := exec.Command("python3", gpath).Output()
	if err != nil {
		out, err = exec.Command("python", gpath).Output()
		if err != nil {
			return err
		}
	}

	if string(out) == "False\n" {
		return errors.New("generate_data.py output False")
	} else if string(out) != "True\n" {
		return errors.New(fmt.Sprintf("generate_data.py uknown output: %v", string(out)))
	}

	return nil
}

func GetVidInfo(path string, filename string, tempjson string, datagen string, tempdata string) (vd.Vidinfo, error) {
	var (
		wg sync.WaitGroup
		vi vd.Vidinfo
	)

	// Geting data about video
	wg.Add(1)
	infob, err := GetMediaInfoJson(path+filename, &wg)
	if err != nil {
		lp.WLog("Error: could not get json data from file")
		removeFile(path, filename)
		return vi, err
	}
	wg.Wait()

	// Writing data to temporary json file
	var raw map[string]interface{}
	json.Unmarshal(infob, &raw)
	info, err := json.Marshal(raw)
	if err != nil {
		lp.WLog("Error: failed to marshal json file")
		removeFile(path, filename)
		return vi, err
	}
	err = ioutil.WriteFile(tempjson, info, 0666)
	if err != nil {
		lp.WLog("Error: could not create json file")
		removeFile(path, filename)
		return vi, err
	}

	// Run python file to get nesessary data from json file
	gpath, err := filepath.Abs(datagen)
	wg.Add(1)
	err = generateDataFile(&wg, gpath)
	wg.Wait()
	if err != nil {
		log.Println(err)
		lp.WLog("Error: failed to generate video data")
		removeFile(path, filename)
		return vi, err
	}

	// Write data to Vidinfo struct
	vi, err = ParseFile(tempdata)
	if err != nil || vi.IsEmpty() {
		log.Println(err)
		lp.WLog("Error: failed parsing data file")
		removeFile(path, filename)
		return vi, err
	}

	return vi, nil
}
