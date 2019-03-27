package transcoder

type (
	Video struct {
		FileName  string
		VtId      int
		VtCodec   string
		FrameRate float64
		VtRes     string
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
		stId     int
		Language string
	}
)
