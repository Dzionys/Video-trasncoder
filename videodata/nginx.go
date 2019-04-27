package vd

type Sequences struct {
	Sequences []Clips `json:"sequences"`
}

type Clips struct {
	Clips []Clip `json:"clips"`
}

type Clip struct {
	Type string `json:"type"`
	Path string `json:"path"`
}
