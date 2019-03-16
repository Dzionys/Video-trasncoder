package transcoder

type videotrack struct {
	index      int
	duration   string
	width      int
	height     int
	frameRate  float64
	codecName  string
	fieldOrder string
}

type audiotrack struct {
	index      int
	channels   int
	sampleRate int
	language   string
	bitRate    int
	codecName  string
}

type subtitle struct {
	index    int
	language string
}

type Vidinfo struct {
	videotracks int
	audiotracks int
	subtitles   int
	videotrack  []videotrack
	audiotrack  []audiotrack
	subtitle    []subtitle
}
