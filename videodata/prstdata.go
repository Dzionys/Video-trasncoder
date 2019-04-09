package vd

type (
	PData struct {
		FileName string
		Streams  []Streams
	}
	Streams struct {
		VtId      string
		VidPreset string
		AudPreset string
		AudioT    []AudT
		subtitleT []SubT
	}
	AudT struct {
		AtId int
	}
	SubT struct {
		AtId int
	}
)

type Preset struct {
	Resolution string
	Codec      string
	Bitrate    int
}
