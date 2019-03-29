package db

import (
	"database/sql"
	"fmt"
	"log"

	transcoder "../transcode"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB          *sql.DB
	tables      tableQueries
	videovalues = []string{"Stream_Id", "Name", "State", "Video_Codec", "Width", "Height", "Frame_Rate"}
	audiovalues = []string{"Stream_Id", "Channels", "Language", "Audio_Codec", "Video_Id"}
	subvalues   = []string{"Stream_Id", "Language", "Video_Id"}
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

func RemoveVideo(name string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
	)

	query = fmt.Sprintf("DELETE FROM Video WHERE Name='%v'", name)
	statement, err = DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func InsertVideo(vid transcoder.Vidinfo, name string, state string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
		id        int
	)

	// Insert video track
	query = getInsertQ(videovalues, "Video")
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
	log.Println(id)

	// Insert audio tracks
	query = getInsertQ(audiovalues, "Audio")
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
	query = getInsertQ(subvalues, "Subtitle")
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

func getInsertQ(clmns []string, tname string) string {
	query := fmt.Sprintf("INSERT INTO %v (", tname)
	val := "("
	for i, c := range clmns {
		if i != len(clmns)-1 {
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

// Not done
func selectQ(clmns []string, tname string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT %v FROM %v", clmns, tname)
	rows, err := DB.Query(query)
	if err != nil {
		return rows, err
	}

	return rows, nil
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
	statement.Exec()

	return nil
}
