package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	vd "../videodata"
	_ "github.com/mattn/go-sqlite3"

	cf "../conf"
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

func RemoveRowByName(name string, tname string) error {
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

func RemoveVideo(name string, stream bool) error {
	var (
		err   error
		path  string
		state string
	)

	CONF, err := cf.GetConf()
	if err != nil {
		return err
	}
	path = CONF.SD

	if stream {
		svnames, err := GetAllStreamVideos(name)
		if err != nil {
			return err
		}
		err = RemoveRowByName(name, "Stream")
		if err != nil {
			return err
		}
		for _, n := range svnames {
			err = os.Remove(path + n)
			if err != nil {
				return err
			}
		}
	} else {
		clms := []string{
			"State",
		}
		query := getSelectQuery(clms, "Video", fmt.Sprintf("Name='%v'", name))
		rows, err := DB.Query(query)
		if err != nil {
			return err
		}
		for rows.Next() {
			rows.Scan(&state)
		}

		if state == "Transcoded" {
			path = CONF.DD
		}

		err = os.Remove(path + name)
		if err != nil {
			return err
		}
		err = RemoveRowByName(name, "Video")
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateVideoName(upname string, oname string, stream bool) error {
	var (
		err       error
		statement *sql.Stmt
		query     string
		tname     string
	)

	if stream {
		tname = "Stream"
	} else {
		tname = "video"

		var (
			state string
			path  string
		)

		CONF, err := cf.GetConf()
		if err != nil {
			return err
		}

		clms := []string{
			"State",
		}
		query = getSelectQuery(clms, "Video", fmt.Sprintf("Name='%v'", oname))
		rows, err := DB.Query(query)
		if err != nil {
			return err
		}
		for rows.Next() {
			rows.Scan(&state)
		}

		if state == "Transcoded" {
			path = CONF.DD
		} else {
			path = CONF.SD
		}

		oldp := path + oname
		newp := path + upname
		err = os.Rename(oldp, newp)
		if err != nil {
			return err
		}
	}
	clms := []string{
		fmt.Sprintf("Name='%v'", upname),
	}
	query = getUpdateQuery(clms, tname, fmt.Sprintf("Name='%v'", oname))
	statement, err = DB.Prepare(query)
	if err != nil {
		println(query)
		return err
	}
	statement.Exec()

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

func getVideoData(vdname string) (vd.Video, string, error) {
	var (
		video vd.Video
		err   error
		query string
		rows  *sql.Rows
		state string
	)

	id, err := getIdByName("Video", vdname)
	if err != nil {
		return video, "", err
	}
	clms := []string{
		"Stream_Id",
		"Channels",
		"Language",
		"Audio_Codec",
	}
	key := fmt.Sprintf("Video_Id=%v", id)
	query = getSelectQuery(clms, "Audio", key)
	rows, err = DB.Query(query)
	if err != nil {
		return video, "", err
	}

	for rows.Next() {
		var tempaud vd.Audio
		rows.Scan(
			&tempaud.AtId,
			&tempaud.Channels,
			&tempaud.Language,
			&tempaud.AtCodec,
		)
		video.AudioT = append(video.AudioT, tempaud)
	}

	clms2 := []string{
		"Stream_Id",
		"Language",
	}
	query = getSelectQuery(clms2, "Subtitle", key)
	rows, err = DB.Query(query)
	if err != nil {
		return video, "", err
	}

	for rows.Next() {
		var tempsub vd.Sub
		rows.Scan(
			&tempsub.StId,
			&tempsub.Language,
		)
		video.SubtitleT = append(video.SubtitleT, tempsub)
	}

	clms3 := []string{
		"Stream_Id",
		"State",
		"Video_Codec",
		"Width",
		"Height",
		"Frame_Rate",
	}
	key2 := fmt.Sprintf("Name='%v'", vdname)
	query = getSelectQuery(clms3, "Video", key2)
	rows, err = DB.Query(query)
	if err != nil {
		return video, "", err
	}

	var (
		width  int
		height int
	)
	for rows.Next() {
		rows.Scan(
			&video.VtId,
			&state,
			&video.VtCodec,
			&width,
			&height,
			&video.FrameRate,
		)
		video.VtRes = strconv.Itoa(width) + "x" + strconv.Itoa(height)
		video.FileName = vdname
	}

	return video, state, nil
}

func GetAllStreamVideos(sname string) ([]string, error) {
	var (
		err   error
		names []string
		query string
		rows  *sql.Rows
	)
	id, err := getIdByName("Stream", sname)
	if err != nil {
		return names, err
	}

	clms := []string{
		"Name",
	}
	key := fmt.Sprintf("Str_Id=%v", id)
	query = getSelectQuery(clms, "Video", key)
	rows, err = DB.Query(query)
	if err != nil {
		return names, err
	}

	var name string
	for rows.Next() {
		rows.Scan(&name)
		names = append(names, name)
	}

	return names, nil
}

func PutVideosToJson() (vd.Dt, error) {
	var (
		videos vd.Dt
		err    error
		query  string
		rows   *sql.Rows
	)

	clms := []string{
		"Name",
	}
	query = getSelectQuery(clms, "Stream", "")
	rows, err = DB.Query(query)
	if err != nil {
		return videos, err
	}
	var streamnames []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		streamnames = append(streamnames, name)
	}

	// Video streams
	for _, n := range streamnames {
		var strvid []string
		strvid, err = GetAllStreamVideos(n)
		if err != nil {
			return videos, err
		}
		var (
			state string
			vid   vd.Video
		)
		var tempvideo vd.VideoStream
		for _, v := range strvid {
			vid, state, err = getVideoData(v)
			if err != nil {
				return videos, err
			}
			tempvideo.Video = append(tempvideo.Video, vid)
		}
		tempvideo.StreamName = n
		tempvideo.State = state
		tempvideo.Stream = true
		videos.VideoStream = append(videos.VideoStream, tempvideo)
	}

	// Indvidual videos
	key := "Str_Id IS NULL"
	query = getSelectQuery(clms, "Video", key)
	rows, err = DB.Query(query)
	if err != nil {
		return videos, err
	}

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	for _, n := range names {
		vid, state, err := getVideoData(n)
		if err != nil {
			return videos, err
		}
		var tempvideo vd.VideoStream
		tempvideo.Stream = false
		tempvideo.State = state
		tempvideo.Video = append(tempvideo.Video, vid)
		videos.VideoStream = append(videos.VideoStream, tempvideo)
	}

	return videos, err
}
