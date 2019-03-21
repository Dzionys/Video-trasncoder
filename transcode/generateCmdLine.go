package transcoder

import (
	"fmt"
	"strconv"
	"strings"
)

// Not tested!!!
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

		// Change resolution
		if crdata.VtRes != svtres {
			vpipe = "scaled"
			maps += "-map [scaled]"
			filter_complex += fmt.Sprintf("[0:v]scale=%v:%v[%v];", res[0], res[1], vpipe)

		} else {
			maps = "-map [v] -map [a]"
		}

		// Change frame rate
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

		// Combine all "-filter_copmplex" filters
		vcode += fmt.Sprintf("%v -filter_complex %v %v ", fps, filter_complex, maps)

	} else {
		mapping += fmt.Sprintf(" -map 0:%v", crdata.VtId)
		for _, at := range crdata.AudioT {
			mapping += " -map 0:" + strconv.Itoa(at.AtId)
		}
	}

	// Changes video codec
	if crdata.VtCodec != "none" {
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
					acode += fmt.Sprintf(bline, cAt.AtId, cAt.Channels*64, cAt.Language)

				} else {
					acode += channels
				}
			}
		}
	}

	// Subtitle part ------------------------------------------

	for _, st := range crdata.SubtitleT {
		scode += fmt.Sprintf(" -c:s:%[1]v copy -metadata:s:s:%[1]v language=%[2]v", st.stId, st.Language)
	}

	cmd = fmt.Sprintf("ffmpeg -i %v %v %v %v %v %v -async 1 -vsync 1 %v", sfname, debugIntr, acode, vcode, scode, mapping, dfname)

	return cmd, nil
}
