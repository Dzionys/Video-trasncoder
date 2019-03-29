package vd

import "reflect"

type videotrack struct {
	Index       int     `json:"index"`
	Duration    string  `json:"duration"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FrameRate   float64 `json:"frameRate"`
	CodecName   string  `json:"codecName"`
	AspectRatio string  `json:"aspectRatio"`
	FieldOrder  string  `json:"fieldOrder"`
}

type audiotrack struct {
	Index      int    `json:"index"`
	Channels   int    `json:"channels"`
	SampleRate int    `json:"sampleRate"`
	Language   string `json:"language"`
	BitRate    int    `json:"bitRate"`
	CodecName  string `json:"codecName"`
}

type subtitle struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
}

type Vidinfo struct {
	Videotracks int          `json:"videotracks"`
	Audiotracks int          `json:"audiotracks"`
	Subtitles   int          `json:"subtitles"`
	Videotrack  []videotrack `json:"videotrack"`
	Audiotrack  []audiotrack `json:"audiotrack"`
	Subtitle    []subtitle   `json:"subtitle "`
}

func (s Vidinfo) IsEmpty() bool {
	return reflect.DeepEqual(s, Vidinfo{})
}
