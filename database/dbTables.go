package db

import (
	"github.com/BurntSushi/toml"
)

type (
	tableQueries struct {
		VideoTable    string
		AudioTable    string
		SubtitleTable string
		PresetTable   string
		PresetValues  [][]string
	}
)

func upTables() (tableQueries, error) {
	var tables tableQueries
	if _, err := toml.DecodeFile("database/tables.toml", &tables); err != nil {
		return tables, err
	}

	return tables, nil
}
