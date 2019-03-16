package transcoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func GetMediaInfoTxt(source string) (string, error) {
	var (
		err error
		wg  sync.WaitGroup
	)
	wg.Add(1)
	infob, err := GetMediaInfoJson(source, &wg)
	if err != nil {
		return "", err
	}
	wg.Wait()

	var raw map[string]interface{}
	json.Unmarshal(infob, &raw)
	info, _ := json.Marshal(raw)
	err = ioutil.WriteFile("transcode/temp.json", info, 0666)
	if err != nil {
		return "", err
	}

	gpath, err := filepath.Abs("transcode/generate_data.py")
	wg.Add(1)
	err = generateDataFile(&wg, gpath)
	wg.Wait()
	if err != nil {
		return "", err
	}

	file, err := ioutil.ReadFile("transcode/temp.txt")
	defer os.Remove("transcode/temp.txt")
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func generateDataFile(wg *sync.WaitGroup, gpath string) error {
	defer wg.Done()

	out, err := exec.Command("python", gpath).Output()
	if err != nil {
		return err
	}

	if string(out) == "False\n" {
		return errors.New("generate_data.py output False")
	} else if string(out) != "True\n" {
		return errors.New(fmt.Sprintf("generate_data.py uknown output: %v", string(out)))
	}

	return nil
}
