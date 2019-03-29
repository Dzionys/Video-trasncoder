package transcoder

import (
	"encoding/json"
	"io/ioutil"
	"os"

	vd "../videodata"
)

func ParseFile(f string) (vd.Vidinfo, error) {
	var (
		vi vd.Vidinfo
	)
	file, err := os.Open(f)
	if err != nil {
		return vi, err
	}
	defer file.Close()
	defer os.Remove(f)

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
