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
