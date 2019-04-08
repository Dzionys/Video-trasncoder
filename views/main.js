'use strict';

function SSE() {
  var source = new EventSource('/sse/dashboard');
  console.log("Connection to /sse/dashboard established")
  var logg = '';
  var currentmsg = '';

  source.onmessage = function(event) {
    if (!event.data.startsWith('<')) {
      logg += '<span class="user">user@transcoder</span>:<span class="home">~</span>$ video-transcode ' + event.data + '<br>';
      localStorage.setItem('filename', event.data)
      document.getElementById('filename').innerText = event.data;
    } else if (event.data.indexOf('Error') > -1) {
      logg += '<span class="error">' + event.data + '</span><br>';
    } else if (/^[\s\S]*<br>.*?Progress:.*?<br>$/.test(logg) && event.data.includes('Progress:')) {
      logg = logg.replace(/^([\s\S]*<br>)(.*?Progress:.*?)(<br>)$/, `$1${event.data}$3`);
    } else {
      currentmsg = event.data;
      logg += currentmsg + '<br>';
    }

    document.getElementById('console').innerHTML = logg;
  };
}

window.onload = SSE();

var formGroupCount = 1;

var inputFile = document.getElementById('input-file');
var transcodeSubmit = document.getElementById('transcode-submit');
var uploadForm = document.getElementById('upload-form');
var transcodeForm = document.getElementById('transcode');
var uploadFormLabel = document.getElementById('upload-form-label');
var uploadFormInput = document.getElementById('input-file');

var codec = document.getElementById('codec');
var resolution = document.getElementById('resolution');
var framerate = document.getElementById('frame-rate');
var audioTracks = document.getElementById('audio-tracks');
var subtitleTracks = document.getElementById('subtitle-tracks');
var formGroup = document.getElementsByClassName('form-group')[0];
var buttonAdd = document.getElementById('button-add');
var add = document.getElementById('form-group-add');

function addFormGroup(add, formGroup, i) {
  let newFormGroup = formGroup.cloneNode(true);

  let label = newFormGroup.getElementsByClassName('vidpre-label')[0];
  let select = newFormGroup.getElementsByClassName('video-presets')[0];
  label.id = `vidpre-label-${i}`;
  select.id = `video-presets-${i}`;
  label.setAttribute('for', `vidpre-${i}`);
  select.setAttribute('name', `vidpre-${i}`);
  
  label = newFormGroup.getElementsByClassName('audpre-label')[0];
  select = newFormGroup.getElementsByClassName('audio-presets')[0];
  label.id = `audpre-label-${i}`;
  select.id = `audio-presets-${i}`;
  label.setAttribute('for', `audpre-${i}`);
  select.setAttribute('name', `audpre-${i}`);

  label = newFormGroup.getElementsByClassName('audioselect-label')[0];
  select = newFormGroup.getElementsByClassName('audio-select')[0];
  label.id = `audioselect-label-${i}`;
  select.id = `audio-select-${i}`;
  label.setAttribute('for', `audioselect-${i}`);
  select.setAttribute('name', `audioselect-${i}`);

  label = newFormGroup.getElementsByClassName('subselect-label')[0];
  select = newFormGroup.getElementsByClassName('subtitle-select')[0];
  label.id = `subselect-label-${i}`;
  select.id = `subtitle-select-${i}`;
  label.setAttribute('for', `subselect-${i}`);
  select.setAttribute('name', `subselect-${i}`);

  add.appendChild(newFormGroup);
}

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
      uploadFormInput.disabled = true;
      resolution.innerText = `${response.data['videotrack'][0]['width']}x${response.data['videotrack'][0]['height']}`;
      codec.innerText = response.data['videotrack'][0]['codecName'];
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

function transcode(event) {
  var video = JSON.parse(localStorage.getItem('video'));
  var data = JSON.parse(localStorage.getItem('cldata'));

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
      // handle error
    });

  event.preventDefault();
}

buttonAdd.addEventListener('click', function(event) {
  formGroupCount++;
  addFormGroup(add, formGroup, formGroupCount);
  event.preventDefault();
});

inputFile.addEventListener('change', upload);
transcodeForm.addEventListener('submit', transcode);
