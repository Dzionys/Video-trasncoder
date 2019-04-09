package transcoder

type Config struct {
	SD         string
	TD         string
	DD         string
	Debug      bool
	DebugStart string
	DebugEnd   string
	TempJson   string
	TempTxt    string
	VBW        int
	ABW        int
	DataGen    string
	LogP       string
	Advanced   bool
	Presets    bool
	FileTypes  []string
}
