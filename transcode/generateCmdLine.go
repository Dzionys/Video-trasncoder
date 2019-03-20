package transcoder

import (
	"fmt"
	"strconv"
	"strings"
)

// Not done yet
func genCmdLine(crdata Video, vdata Vidinfo, sfname string, dfname string) (string, error) {
	var (
		cmd       = ""
		mapping   = ""
		vcode     = ""
		acode     = ""
		scode     = ""
		debugIntr = ""
	)

	// Checks if debuging is set to true
	if DEBUG {
		debugIntr = "-ss " + CONF.DebugStart + " -t " + CONF.DebugEnd
	}

	// Video part ---------------------------------------------

	// Change frames per second and/or resolution
	svtres := strconv.Itoa(vdata.Videotrack[0].Width) + ":" + strconv.Itoa(vdata.Videotrack[0].Height)
	if crdata.FrameRate != vdata.Videotrack[0].FrameRate || crdata.VtRes != svtres {

		res := strings.Split(crdata.VtRes, ":")
		var (
			filter_complex = ""
			vpipe          = "0:v"
			fps            = ""
			maps           = ""
		)

		if crdata.VtRes != svtres {
			vpipe = "scaled"
			maps += "-map [scaled]"
			filter_complex += fmt.Sprintf("[0:v]scale=%v:%v[%v];", res[0], res[1], vpipe)
		} else {
			maps = "-map [v] -map [a]"
		}

		if crdata.FrameRate != vdata.Videotrack[0].FrameRate {
			fps = fmt.Sprintf(" -r %v", crdata.FrameRate)
			bline := "[%[3]v]setpts=%[2]v/%[1]v*PTS[v];[0:a]atempo=%[1]v/%[2]v[a]"
			filter_complex += fmt.Sprintf(bline, crdata.FrameRate, vdata.Videotrack[0].FrameRate, vpipe)
		} else {

			// Map all audio tracks if not mapped while changing fps
			for _, at := range crdata.AudioT {
				mapping += " -map 0:" + strconv.Itoa(at.AtId)
			}
		}

		vcode += fmt.Sprintf("%v -filter_complex %v %v ", fps, filter_complex, maps)

	} else {
		mapping += fmt.Sprintf(" -map 0:%v", crdata.VtId)
		for _, at := range crdata.AudioT {
			mapping += " -map 0:" + strconv.Itoa(at.AtId)
		}
	}

	// Audio part ---------------------------------------------

	for _, at := range crdata.AudioT {
		acode += fmt.Sprintf(" -c:a:%[1]v libfdk_aac -ac 2 -b:a:%[1]v %[2]vk -metadata language=%[3]v", at.AtId, CONF.ABW, at.Language)
	}

	// Subtitle part ------------------------------------------

	for _, st := range crdata.SubtitleT {
		scode += fmt.Sprintf(" -c:s:%[1]v copy -metadata:s:s:%[1]v language=%[2]v", st.stId, st.Language)
	}

	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v -async 1 -vsync 1 %v", sfname, debugIntr, vcode, acode, scode, mapping, dfname)

	return cmd, nil
}
