package transcoder

import (
	"fmt"
	"strconv"
	"strings"

	"../lp"
	vd "../videodata"
)

// Needs more testing !!!
func generateClientCmdLine(crdata vd.Video, vdata vd.Vidinfo, sf string, sfname string, df string) (string, error) {
	var (
		cmd       = ""
		mapping   = ""
		vcode     = ""
		acode     = ""
		scode     = ""
		debugIntr = ""
	)

	// Checks if debuging is set to true
	if CONF.Debug {
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

		// Change resolution
		if crdata.VtRes != svtres {
			vpipe = "scaled"
			maps += "-map [scaled]"
			filter_complex += fmt.Sprintf("[0:v]scale=%v:%v[%v]", res[0], res[1], vpipe)
		}

		// Change frame rate
		if crdata.FrameRate != vdata.Videotrack[0].FrameRate {
			maps = "-map [v] -map [a]"
			fps = fmt.Sprintf(" -r %v", crdata.FrameRate)
			var bline string
			if vpipe == "scaled" {
				bline = ";[%[3]v]setpts=%[2]v/%[1]v*PTS[v];[0:a]atempo=%[1]v/%[2]v[a]"
			} else {
				bline = "[%[3]v]setpts=%[2]v/%[1]v*PTS[v];[0:a]atempo=%[1]v/%[2]v[a]"
			}
			filter_complex += fmt.Sprintf(bline, crdata.FrameRate, vdata.Videotrack[0].FrameRate, vpipe)

		} else {

			// Map all audio tracks if not mapped while changing fps
			for _, at := range crdata.AudioT {
				mapping += " -map 0:" + strconv.Itoa(at.AtId)
			}
		}

		// Combine all "-filter_copmplex" filters
		vcode += fmt.Sprintf("%v -filter_complex %v %v ", fps, filter_complex, maps)

	} else {
		mapping += fmt.Sprintf(" -map 0:%v", crdata.VtId)
		for _, at := range crdata.AudioT {
			mapping += " -map 0:" + strconv.Itoa(at.AtId)
		}
	}

	// Changes video codec
	if crdata.VtCodec != "nochange" {
		switch crdata.VtCodec {

		case "h264":
			vcode += fmt.Sprintf(" -c:v:%[1]v libx264 -b:v:%[1]v %[2]vk -metadata:s:v:%[1]v name=\"%[3]v\"", crdata.VtId, CONF.VBW, sfname)
			break

		case "h265":
			templ := fmt.Sprintf(" -c:v:%v libx265 -x265-params \"preset=slower:me=hex:no-rect=1:no-amp=1:rd=4:aq-mode=2:", crdata.VtId)
			templ += "aq-strength=0.5:psy-rd=1.0:psy-rdoq=0.2:bframes=3:min-keyint=1\" "
			templ += fmt.Sprintf("-b:v:0 %vk -metadata:s:v:0 name=\"%v\"", CONF.VBW, sfname)
			vcode += templ
		}
	} else {
		vcode += fmt.Sprintf(" -c:v:%[1]v copy -metadata:s:v:%[1]v name=\"%[2]v\"", crdata.VtId, sfname)
	}

	// Audio part ---------------------------------------------

	for _, cAt := range crdata.AudioT {
		for _, sAt := range vdata.Audiotrack {
			if cAt.AtId == sAt.Index {

				channels := ""
				bline := " -c:a:%[1]v libfdk_aac%[4]v -b:a:%[1]v %[2]vk -metadata language=%[3]v"

				// If frame rates changed do not map
				if !(crdata.FrameRate != vdata.Videotrack[0].FrameRate) {
					mapping += fmt.Sprintf(" -map 0:%v", cAt.AtId)
				}

				// Change layout to stereo or mono
				if cAt.Channels != sAt.Channels {
					switch cAt.Channels {

					case 2:
						channels = " -ac 2"

					case 1:
						channels = " -ac 1"
					}
				}

				// Change audio codec to aac
				if cAt.AtCodec != sAt.CodecName {
					acode += fmt.Sprintf(bline, cAt.AtId, cAt.Channels*64, cAt.Language, channels)

				} else {
					acode += channels
				}
			}
		}
	}

	// Subtitle part ------------------------------------------

	for _, st := range crdata.SubtitleT {
		scode += fmt.Sprintf(" -c:s:%[1]v copy -metadata:s:s:%[1]v language=%[2]v", st.StId, st.Language)
	}

	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v -async 1 -vsync 1 %v", sf, debugIntr, acode, vcode, scode, mapping, df)

	return cmd, nil
}

func generateBaseCmdLine(d vd.Vidinfo, sf string, df string, sfname string) string {
	var (
		cmd     string
		mapping string
		vcode   string
		acode   string
		scode   string
		frate   string
		ss      string
	)

	// Video cmd
	if d.Videotrack[0].FrameRate < 25 {
		baseln := "-r 25 -filter_complex [0:v]setpts=%v/25*PTS[v];[0:a]atempo=25/%v[a] -map [v] -map [a]"
		frate = fmt.Sprintf(baseln, d.Videotrack[0].FrameRate, d.Videotrack[0].FrameRate)
	} else {
		frate = ""
		mapping = fmt.Sprintf("-map 0:%v", d.Videotrack[0].Index)
	}

	if d.Videotrack[0].Width > 1280 || d.Videotrack[0].CodecName != "h265" {
		vcode = "-c:v:0 libx265 -x265-params \"preset=slower:me=hex:no-rect=1:no-amp=1:rd=4:aq-mode=2:"
		vcode += "aq-strength=0.5:psy-rd=1.0:psy-rdoq=0.2:bframes=3:min-keyint=1\" "
		vcode += fmt.Sprintf("-b:v:0 %vk -metadata:s:v:0 name=\"%v\"", CONF.VBW, sfname)
		if d.Videotrack[0].Width > 1280 {
			vcode += " -filter:v:0 \"scale=iw*sar:ih,pad=iw:iw/16*9:0:(oh-ih)/2\" -aspect 16:9"
		}
	} else {
		vcode = fmt.Sprintf(" -c:v:0 copy -metadata:s:v:0 name=\"%v\"", sfname)
	}

	// Audio cmd
	acode = ""
	for i := 0; i < d.Audiotracks; i++ {
		if !(d.Videotrack[0].FrameRate < 25) {
			mapping += fmt.Sprintf(" -map 0:%v", d.Audiotrack[i].Index)
		}
		acode += fmt.Sprintf(" -c:a:%v libfdk_aac -ac 2 -b:a:%v %vk -metadata language=%v", i, i, CONF.ABW, d.Audiotrack[i].Language)
	}

	// Subtitles cmd
	scode = ""
	if d.Subtitles > 0 {
		for i := 0; i < d.Subtitles; i++ {
			scode += fmt.Sprintf(" -c:s:%v copy -metadata:s:s:%v language=%v", d.Subtitle[i].Index, d.Subtitle[i].Index, d.Subtitle[i].Language)
		}
	}

	// Set debug duration if debug is selected
	ss = ""
	if CONF.Debug {
		ss = "-ss " + CONF.DebugStart + " -t " + CONF.DebugEnd
	}

	// Add all parts in one command line
	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v %v -async 1 -vsync 1 %v", sf, ss, mapping, frate, vcode, acode, scode, df)
	lp.WLog("Command line generated")

	return cmd
}
