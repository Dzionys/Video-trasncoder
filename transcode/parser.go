package transcoder

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func ParseFile(f string) (Vidinfo, error) {
	var (
		vi Vidinfo
		ln []string
	)
	file, err := os.Open(f)
	if err != nil {
		return vi, err
	}
	defer os.Remove(f)
	scanner := bufio.NewScanner(file)

	scanner.Scan()
	ln = strings.Split(scanner.Text(), " ")
	vi.videotracks, _ = strconv.Atoi(ln[1])
	scanner.Scan()
	ln = strings.Split(scanner.Text(), " ")
	vi.audiotracks, _ = strconv.Atoi(ln[1])
	scanner.Scan()
	ln = strings.Split(scanner.Text(), " ")
	vi.subtitles, _ = strconv.Atoi(ln[1])

	vi.videotrack = make([]videotrack, vi.videotracks)
	for i := 0; i < vi.videotracks; i++ {
		scanner.Scan()
		ln = strings.Split(scanner.Text(), " ")
		if ln[0] == "videotrack" {
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].index, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].width, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].height, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].frameRate, _ = strconv.ParseFloat(ln[1], 64)
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].codecName = ln[1]
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.videotrack[i].fieldOrder = ln[1]
		}
	}
	vi.audiotrack = make([]audiotrack, vi.audiotracks)
	for i := 0; i < vi.audiotracks; i++ {
		scanner.Scan()
		ln = strings.Split(scanner.Text(), " ")
		if ln[0] == "audiotrack" {
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].index, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].channels, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].sampleRate, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].language = ln[1]
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].bitRate, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.audiotrack[i].codecName = ln[1]
		}
	}
	vi.subtitle = make([]subtitle, vi.subtitles)
	for i := 0; i < vi.subtitles; i++ {
		scanner.Scan()
		ln = strings.Split(scanner.Text(), " ")
		if ln[0] == "subtitle" {
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.subtitle[i].index, _ = strconv.Atoi(ln[1])
			scanner.Scan()
			ln = strings.Split(scanner.Text(), " ")
			vi.subtitle[i].language = ln[1]
		}
	}

	return vi, nil
}
