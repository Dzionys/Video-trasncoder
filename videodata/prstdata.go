package vd

type (
	PData struct {
		FileName string
		Save     bool
		Streams  []Streams
	}
	Streams struct {
		VtId      int
		VidPreset string
		AudPreset string
		AudioT    []AudT
		SubtitleT []SubT
	}
	AudT struct {
		AtId int
		Lang string
	}
	SubT struct {
		StId int
		Lang string
	}
)

type Preset struct {
	Resolution string
	Codec      string
	Bitrate    int
}
