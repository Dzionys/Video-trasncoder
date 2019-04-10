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
	streamvalues = []string{"Name"}
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
	err = prepareTable(tables.StreamTable)
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

func RemoveColumnByName(name string, tname string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
	)

	query = getDeleteQuery(tname, fmt.Sprintf("Name='%v'", name))
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

func InsertStream(vids []vd.Vidinfo, names []string, state string, sname string) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
		id        int
	)

	// Insert new stream
	query = getInsertQuery(streamvalues, "Stream")
	statement, err = DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = statement.Exec(sname)

	// Get stream id
	id, err = getIdByName("Stream", sname)
	if err != nil {
		return err
	}

	// Insert videos into stream
	for i, vid := range vids {
		InsertVideo(vid, names[i], state, id)
	}

	return nil
}

func InsertVideo(vid vd.Vidinfo, name string, state string, sid int) error {
	var (
		err       error
		query     string
		statement *sql.Stmt
		id        int
	)

	// Insert video track
	if sid < 0 {
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
	} else {
		vidval := videovalues
		vidval = append(vidval, "Str_Id")
		query = getInsertQuery(vidval, "Video")
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
			sid,
		)
	}
	if err != nil {
		return err
	}

	// Get video id using as foreign keys in audio and video tracks
	id, err = getIdByName("Video", name)
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

func AddPresetsToJson(vid vd.Vidinfo) (vd.Data, error) {
	var (
		data  vd.Data
		query string
		err   error
		rows  *sql.Rows
	)
	var (
		name       string
		resolution string
		codec      string
		bitrate    int
	)

	clms := []string{
		"Name",
		"Resolution",
		"Codec",
		"Bitrate",
	}

	query = getSelectQuery(clms, "Preset", "Type=0")
	rows, err = DB.Query(query)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		rows.Scan(&name, &resolution, &codec, &bitrate)
		temp := vd.Videopresets{
			name,
			resolution,
			codec,
			bitrate,
		}
		data.Vidpresets = append(data.Vidpresets, temp)
	}

	query = getSelectQuery(clms, "Preset", "Type=1")
	rows, err = DB.Query(query)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		rows.Scan(&name, &resolution, &codec, &bitrate)
		temp := vd.Audiopresets{
			name,
			codec,
			bitrate,
		}
		data.Audpresets = append(data.Audpresets, temp)
	}
	data.Vidinfo = vid

	return data, nil
}

func GetPreset(name string) (vd.Preset, error) {
	var (
		prst  vd.Preset
		query string
		err   error
		rows  *sql.Rows
	)

	clms := []string{
		"Resolution",
		"Codec",
		"Bitrate",
	}

	query = getSelectQuery(clms, "Preset", fmt.Sprintf("Name='%v'", name))
	rows, err = DB.Query(query)
	if err != nil {
		return prst, err
	}

	for rows.Next() {
		rows.Scan(&prst.Resolution, &prst.Codec, &prst.Bitrate)
	}

	return prst, nil
}
