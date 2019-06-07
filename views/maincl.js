'use strict';

window.onload = sse();

var save = false;

function upload(event) {
  var file = event.target.files[0];
  if (!file) {
    return;
  }

  var formData = new FormData();
  formData.append('file', file);
  axios.post('/upload', formData)
    .then(function (response) {

      uploadForm.className = 'upload-form uploaded';
      uploadFormLabel.className = 'upload-form-label uploaded';
      transcodeForm.className = 'transcode-form uploaded';
      toggle.className = 'toggle uploaded';
      checkBox.disabled = true;
      uploadFormInput.disabled = true;
      resolution.innerText = `${response.data['videotrack'][0]['width']}x${response.data['videotrack'][0]['height']}, `;
      codec.innerText = `${response.data['videotrack'][0]['codecName']}, `;
      framerate.innerText = `${response.data['videotrack'][0]['frameRate']}fps`;

      if (!response.data.audiotracks == 0){
        response.data.audiotrack.map(function (value) {
          audioTracks.innerHTML += `<option selected value="${value.index}">${value.index}, ${value.language}, ${value.channels}ch</option>`
        })
      } else {
        audioTracks.innerHTML += `<option selected value="null">no tracks</option>`
      }

      if (!response.data.subtitles == 0){
        response.data.subtitle.map(function (value) {
          subtitleTracks.innerHTML += `<option selected value="${value.index}">${value.index}, ${value.language}</option>`
        })
      } else {
        subtitleTracks.innerHTML += `<option selected value="null">no tracks</option>`
      }

      // Client data json pattern
      var data = {
        "FileName": localStorage.getItem('filename'),
        "VtId": response.data['videotrack'][0]['Index'],
        "VtCodec": "",
        "FrameRate": 0.0,
        "VtRes": "",
        "Save": false,
        "AudioT": [],
        "SubtitleT": []
      };

      localStorage.removeItem('filename');
      localStorage.setItem('video', JSON.stringify(response.data));
      localStorage.setItem('cldata', JSON.stringify(data));
    
    })
    .catch(function (error) {
      console.log(error)
      // handle error
    });
}

function transcode(event, save) {
  var video = JSON.parse(localStorage.getItem('video'));
  var data = JSON.parse(localStorage.getItem('cldata'));

  data['Save'] = save;

  // Video codec
  var e = document.getElementById('codec-select');
  if (e.options[e.selectedIndex].value != "nochange") {
  data['VtCodec'] = e.options[e.selectedIndex].value;
  } else {
    data['VtCodec'] = video['videotrack'][0]['codecName'];
  }

  // Video resolution
  e = document.getElementById('resolution-select');
  if (e.options[e.selectedIndex].value != "nochange") {
    data['VtRes'] = e.options[e.selectedIndex].value;
  } else {
    data['VtRes'] = `${video['videotrack'][0]['width']}:${video['videotrack'][0]['height']}`;
  }

  // Video frame rate
  e = document.getElementById('fr-select');
  if (e.options[e.selectedIndex].value != "nochange") {
    data['FrameRate'] = parseFloat(e.options[e.selectedIndex].value);
  } else {
    data['FrameRate'] = video['videotrack'][0]['frameRate'];
  }

  // Audio tracks
  e = document.getElementById('audio-tracks');
  var audioindex = e.options[e.selectedIndex].value;
  if (audioindex != "null") {
    var audiocodec = "";
    var chann = 0;
    var e2 = document.getElementById('audio-select');
    var e3 = document.getElementById('channels-select');
    var atindex = video.audiotrack.findIndex(item => item.index == audioindex);

    if (e2.options[e2.selectedIndex].value != "nochange") {
      audiocodec = e.options[e.selectedIndex].value;
    } else {
      audiocodec = video['audiotrack'][atindex]['codecName'];
    }
    if (e3.options[e3.selectedIndex].value != "nochange") {
      chann = e3.options[e3.selectedIndex].value;
    } else {
      chann = video['audiotrack'][atindex]['channels'];
    }
    var audio = {
      "AtId": parseInt(audioindex),
      "AtCodec": audiocodec,
      "Language": video['audiotrack'][atindex]['language'],
      "Channels": parseInt(chann)
    };
    data['AudioT'].push(audio)
  }

  // Subtitle tracks
  e = document.getElementById('subtitle-tracks');
  var subindex = e.options[e.selectedIndex].value;
  if (subindex != "null") {
    var stindex = video.subtitle.findIndex(item => item.index == subindex);
    var sub = {
      "StId": parseInt(subindex),
      "Language": video['subtitle'][stindex]['Language']
    };
    data['SubtitleT'].push(sub)
  }

  // Send client data to server
  axios.post('/transcode', data)
    .then(function (response) {
      localStorage.removeItem('cldata');
      localStorage.removeItem('video')
      transcodeForm.className = 'transcode-form';
    })
    .catch(function (error) {
      console.log(error)
      // handle errortranscodeForm.addEventListener('submit', transcode(event, false));
buttonSave.addEventListener('submit', transcode(event, true));
    });

  event.preventDefault();
}

inputFile.addEventListener('change', upload);
transcodeForm.addEventListener('submit', transcode);
buttonSave.addEventListener('click', function(event) {
  save = true;
  transcode(event, save);
});