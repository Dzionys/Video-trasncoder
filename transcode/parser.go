package transcoder

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ParseFile(f string) (Vidinfo, error) {
	var (
		vi Vidinfo
	)
	file, err := os.Open(f)
	if err != nil {
		return vi, err
	}
	defer file.Close()
	defer os.Remove(f)

	println(file)

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return vi, err
	}

	err = json.Unmarshal(byteValue, &vi)
	if err != nil {
		return vi, err
	}

	return vi, nil
}
