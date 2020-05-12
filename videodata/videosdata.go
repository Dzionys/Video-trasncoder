package vd

type (
	Video struct {
		FileName  string
		VtId      int
		VtCodec   string
		FrameRate float64
		VtRes     string
		Save      bool
		AudioT    []Audio
		SubtitleT []Sub
	}
	Audio struct {
		AtId     int
		AtCodec  string
		Language string
		Channels int
	}
	Sub struct {
		StId     int
		Language string
	}
)

type Dt struct {
	VideoStream []VideoStream
}

type VideoStream struct {
	Stream     bool
	StreamName string
	State      string
	Thumbnail  string
	Video      []Video
}
