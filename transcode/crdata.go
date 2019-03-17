package transcoder

type (
	Video struct {
		FileName string
		VtId int
		VtCodec string
		VtRes string
		AudioT []Audio
		SubtitleT []Sub
	}
	Audio struct {
		AtId int
		AtCodec string
	}
	Sub struct {
		stId int
	}
)