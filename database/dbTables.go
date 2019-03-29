package db

import (
	"github.com/BurntSushi/toml"
)

type (
	tableQueries struct {
		VideoTable    string
		AudioTable    string
		SubtitleTable string
	}
)

func upTables() (tableQueries, error) {
	var tables tableQueries
	if _, err := toml.DecodeFile("db/dbTable.toml", &tables); err != nil {
		return tables, err
	}

	return tables, nil
}
