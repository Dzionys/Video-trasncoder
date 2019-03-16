import json
from pprint import pprint
import re
import os

CDX = {
    'video': ['amv', 'mpeg2video', 'mpeg4', 'msmpeg4v2', 'msmpeg4v3', 'msmpeg4v2', 'h264', 'hevc', 'theora', 'wmv1', 'wmv2', 'vp8', 'vp9'],
    'audio': ['aac', 'ac3', 'dts', 'eac3', 'mp2', 'mp3', 'opus', 'wmav1', 'wmav2', 'vorbis'],
    'subtitles': ['srt','subrip']
}
SAL = ['rus','eng','lit','fra','ger','ita','org']

v = {}
a = {}
u = {}
vc = 0
ac = 0
sc = 0

jsonfile = 'transcode/temp.json'
txtfile = 'transcode/temp.txt'

with open(jsonfile) as f:
    data = json.load(f)

def parseData():
    global v
    global a
    global u
    global vc
    global ac
    global sc

    try:
        for s in data['streams']:
            if s['codec_type'] == 'video':
                v[vc] = {}
                v[vc]['index'] = int(s['index'])
                v[vc]['codec_name'] = s['codec_name']
                if 'width' in s:
                    v[vc]['width'] = int(s['width'])
                else:
                    v[vc]['width'] = 0
                if 'height' in s:
                    v[vc]['height'] = int(s['height'])
                else:
                    v[vc]['height'] = 0
                if 'r_frame_rate' in s:
                    f = s['r_frame_rate'].split('/')
                    fr = float(f[0])
                    sk = float(f[1])
                    fps = fr/sk
                    v[vc]['frame_rate'] = round(fps, 3)
                else:
                    v[vc]['frame_rate'] = 0
                if 'field_order' in s:
                    v[vc]['field_order'] = s['field_order']
                else:
                    v[vc]['field_order'] = ""
                vc = vc+1
            if s['codec_type'] == 'audio':
                a[ac] = {}
                a[ac]['index'] = int(s['index'])
                if 'tags' in s:
                    if 'language' in s['tags']:
                        a[ac]['language'] = s['tags']['language']
                    else:
                        a[ac]['language'] = 'undefined'
                else:
                    a[ac]['language'] = 'undefined'
                if 'codec_name' in s:
                    a[ac]['codec_name'] = s['codec_name']
                else:
                    a[ac]['codec_name'] = ""
                if 'channels' in s:
                    a[ac]['channels'] = int(s['channels'])
                else:
                    a[ac]['channels'] = 0
                if 'sample_rate' in s:
                    a[ac]['sample_rate'] = int(s['sample_rate'])
                else:
                    a[ac]['sample_rate'] = 0
                if 'bit_rate' in s:
                    a[ac]['bit_rate'] = int(s['bit_rate'])
                else:
                    a[ac]['bit_rate'] = 0
                ac = ac+1
            if s['codec_name'] in CDX['subtitles']:
                if 'language' in s['tags']:
                    u[sc] = {}
                    u[sc]['index'] = int(s['index'])
                    u[sc]['lang'] = s['tags']['language']
                    sc = sc+1
                else:
                    if 'title' in s['tags']:
                        cc = '({0}).*'.format('|'.join('{0}'.format(s) for s in SAL))
                        m = re.search(cc,s['tags']['title'].lower())
                        if m:
                            u[sc] = {}
                            u[sc]['index'] = int(s['index'])
                            u[sc]['lang'] = m.group(1)
                    sc = sc+1
        print(True)
        os.remove(jsonfile)
    except Exception as e:
        print(False)
        return

def writeToFile():
    file = open(txtfile, "w")

    file.write("videotracks {0}\n".format(vc))
    file.write("audiotracks {0}\n".format(ac))
    file.write("subtitles {0}\n".format(sc))

    for z in v:
        file.write("videotrack {0}\n".format(z))
        file.write("index {0}\n".format(v[z]["index"]))
        if 'width' in v[z]:
            file.write("width {0}\n".format(v[z]["width"]))
        if 'height' in v[z]:
            file.write("height {0}\n".format(v[z]["height"]))
        if 'frame_rate' in v[z]:
            file.write("frame_rate {0}\n".format(v[z]["frame_rate"]))
        if 'codec_name' in v[z]:
            file.write("codec_name {0}\n".format(v[z]["codec_name"]))
        if 'field_order' in v[z]:
            file.write("field_order {0}\n".format(v[z]["field_order"]))

    for x in a:
        file.write("audiotrack {0}\n".format(x))
        file.write("index {0}\n".format(a[x]["index"]))
        if 'channels' in a[x]:
            file.write("channels {0}\n".format(a[x]["channels"]))
        if 'sample_rate' in a[x]:
            file.write("sample_rate {0}\n".format(a[x]["sample_rate"]))
        if 'language' in a[x]:
            file.write("language {0}\n".format(a[x]["language"]))
        if 'bit_rate' in a[x]:
            file.write("bit_rate {0}\n".format(a[x]["bit_rate"]))
        if 'codec_name' in a[x]:
            file.write("codec_name {0}\n".format(a[x]["codec_name"]))
    
    for c in u:
        file.write("subtitle {0}\n".format(c))
        file.write("index {0}\n".format(u[c]["index"]))
        if 'lang' in u[c]:
            file.write("language {0}\n".format(u[c]["lang"]))

    file.close()

            

def main():
    parseData()
    writeToFile()

if __name__ =='__main__':main()
