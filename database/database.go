package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	vd "../videodata"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB           *sql.DB
	tables       tableQueries
	videovalues  = []string{"Stream_Id", "Name", "State", "Video_Codec", "Width", "Height", "Frame_Rate"}
	audiovalues  = []string{"Stream_Id", "Channels", "Language", "Audio_Codec", "Video_Id"}
	subvalues    = []string{"Stream_Id", "Language", "Video_Id"}
	presetvalues = []string{"Name", "Type", "Resolution", "Codec", "Bitrate"}
)

func OpenDatabase() error {
	var err error

	// Upload tables from tables.toml
	tables, err = upTables()
	if err != nil {
		log.Println("Error: failed to upload database tables")
		log.Println(err)
		return err
	}

	// Open database "data.db" if not exist creates new one
	DB, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Println("Error: failed to open database")
		log.Println(err)
		return err
	}

	// Creates tables if not exist
	err = prepareTable(tables.VideoTable)
	if err != nil {
		return err
	}
	err = prepareTable(tables.AudioTable)
	if err != nil {
		return err
	}
	err = prepareTable(tables.SubtitleTable)
	if err != nil {
		return err
	}

	// Recreate Preset table
	err = runCustomQuery("DROP TABLE IF EXISTS Preset")
	if err != nil {
		return err
	}
	err = prepareTable(tables.PresetTable)
	if err != nil {
		return err
	}

	// Enable foreign keys in database
	statement, err := DB.Prepare("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func UpdateState(name string, state string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
	)

	clms := []string{fmt.Sprintf("State='%v'", state)}
	query = getUpdateQuery(clms, "Video", fmt.Sprintf("Name='%v'", name))
	statement, err = DB.Prepare(query)
	if err != nil {
		log.Println("Error query: " + query)
		return err
	}
	statement.Exec()

	return nil
}

func RemoveVideo(name string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
	)

	query = getDeleteQuery("Video", fmt.Sprintf("Name='%v'", name))
	statement, err = DB.Prepare(query)
	if err != nil {
		log.Println("Error query: " + query)
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func InsertVideo(vid vd.Vidinfo, name string, state string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
		id        int
	)

	// Insert video track
	query = getInsertQuery(videovalues, "Video")
	statement, err = DB.Prepare(query)
	if err != nil {
		log.Println("Error query: " + query)
		return err
	}
	_, err = statement.Exec(
		vid.Videotrack[0].Index,
		name,
		state,
		vid.Videotrack[0].CodecName,
		vid.Videotrack[0].Width,
		vid.Videotrack[0].Height,
		vid.Videotrack[0].FrameRate,
	)
	if err != nil {
		return err
	}

	// Get video id using as foreign keys in audio and video tracks
	id, err = getVidId(name)
	if err != nil {
		return err
	}

	// Insert audio tracks
	query = getInsertQuery(audiovalues, "Audio")
	statement, err = DB.Prepare(query)
	if err != nil {
		log.Println("Error query: " + query)
		return err
	}
	for _, a := range vid.Audiotrack {
		_, err = statement.Exec(
			a.Index,
			a.Channels,
			a.Language,
			a.CodecName,
			id,
		)
		if err != nil {
			return err
		}
	}

	// Insert subtitle tracks
	query = getInsertQuery(subvalues, "Subtitle")
	statement, err = DB.Prepare(query)
	if err != nil {
		log.Println("Error query: " + query)
		return err
	}
	for _, s := range vid.Subtitle {
		_, err = statement.Exec(
			s.Index,
			s.Language,
			id,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertPresets() error {
	query := getInsertQuery(presetvalues, "Preset")

	for _, v := range tables.PresetValues {
		statement, err := DB.Prepare(query)
		if err != nil {
			return err
		}
		tp, err := strconv.Atoi(v[1])
		if err != nil {
			return err
		}
		_, err = statement.Exec(v[0], tp, v[2], v[3], v[4])
		if err != nil {
			return err
		}
	}

	return nil
}

func getInsertQuery(clms []string, tname string) string {
	query := fmt.Sprintf("INSERT INTO %v (", tname)
	val := "("
	for i, c := range clms {
		if i != len(clms)-1 {
			query += fmt.Sprintf("%v,", c)
			val += "?,"
		} else {
			query += fmt.Sprintf("%v", c)
			val += "?"
		}
	}
	val += ")"
	query += fmt.Sprintf(") VALUES %v", val)

	return query
}

func getSelectQuery(clms []string, tname string, key string) string {
	query := "SELECT "
	if len(clms) > 0 {
		for i, c := range clms {
			if i != len(clms)-1 {
				query += fmt.Sprintf("%v,", c)
			} else {
				query += fmt.Sprintf("%v ", c)
			}
		}
	} else {
		query += "* "
	}
	query += "FROM " + tname
	if key != "" {
		query += " WHERE " + key
	}

	return query
}

func getDeleteQuery(tname string, key string) string {
	query := fmt.Sprintf("DELETE FROM %v WHERE %v", tname, key)
	return query
}

func getUpdateQuery(clms []string, tname string, key string) string {
	query := fmt.Sprintf("UPDATE %v SET ", tname)
	for i, c := range clms {
		if i != len(clms)-1 {
			query += fmt.Sprintf("%v,", c)
		} else {
			query += fmt.Sprintf("%v ", c)
		}
	}
	query += "WHERE " + key

	return query
}

func getVidId(name string) (int, error) {
	var id = -1
	row, err := DB.Query(fmt.Sprintf("SELECT Id FROM Video WHERE Name='%v'", name))
	if err != nil {
		return id, err
	}
	for row.Next() {
		err = row.Scan(&id)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func prepareTable(table string) error {
	statement, err := DB.Prepare(table)
	if err != nil {
		log.Println("Error: failed to prepare database table")
		log.Println(err)
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func runCustomQuery(query string) error {
	statement, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}
