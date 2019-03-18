# Video-trasncoder
Video file transcoding system.

### Requiremets
  - ffmpeg
  - golang (used version 1.11.5)
  - python (used version 3.6.7 should work with 2.7.*)

You need to compile ffmpeg with: NASM, Yasm, libx264, libx265 and libfdk-acc in order to be able to use ACC audio encoder.
Full guide for all platforms: http://trac.ffmpeg.org/wiki/CompilationGuide

Your browser must support HTML5 Sever-Sent-Events.

### External golang libraries
  - github.com/BurntSushi/toml

```sh
$ go get github.com/BurntSushi/toml
```

### Run
```sh
$ go run server.go
```
### Open:
```
localhost:8080/
```