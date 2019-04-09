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
      document.getElementById('filename').innerText = `${event.data}, `;
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
  var newFormGroup = formGroup.cloneNode(true);

  var label = newFormGroup.getElementsByClassName('vidpre-label')[0];
  var select = newFormGroup.getElementsByClassName('video-presets')[0];
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
      resolution.innerText = `${response.data['Vidinfo']['videotrack'][0]['width']}x${response.data['Vidinfo']['videotrack'][0]['height']}, `;
      codec.innerText = `${response.data['Vidinfo']['videotrack'][0]['codecName']}, `;
      framerate.innerText = `${response.data['Vidinfo']['videotrack'][0]['frameRate']}fps`;
      
      var vdpr = document.getElementsByClassName('video-presets')[0];
      for (var vp in response.data['Vidpresets']) {
        vdpr.innerHTML += `<option value="${response.data['Vidpresets'][vp]['Name']}">${response.data['Vidpresets'][vp]['Name']}</option>`
      }

      var adpr = document.getElementsByClassName('audio-presets')[0];
      for (var ap in response.data['Audpresets']) {
        adpr.innerHTML += `<option value="${response.data['Audpresets'][ap]['Name']}">${response.data['Audpresets'][ap]['Name']}</option>`
      }

      var adsl = document.getElementsByClassName('audio-select')[0];
      adsl.innerHTML += `<option selected value="keep">keep all</option>`
      if (!response.data['Vidinfo']['audiotracks'] == 0){
        for (var j = 0; j < response.data['Vidinfo']['audiotracks']; j++) {
          adsl.innerHTML += `<option value="${response.data['Vidinfo']['audiotrack'][j]['index']}">${response.data['Vidinfo']['audiotrack'][j]['language']}</option>`
        }
      } else {
        adsl.innerHTML += `<option selected value="null">no tracks</option>`
      }

      var stsl = document.getElementsByClassName('subtitle-select')[0];
      stsl.innerHTML += `<option selected value="keep">keep all</option>`
      if (!response.data['Vidinfo']['subtitles'] == 0){
        for (var j = 0; j < response.data['Vidinfo']['subtitles']; j++) {
          stsl.innerHTML += `<option value="${response.data['Vidinfo']['subtitle'][j]['index']}">${response.data['Vidinfo']['subtitle'][j]['language']}</option>`
        }
      } else {
        stsl.innerHTML += `<option selected value="null">no tracks</option>`
      }

      // Client data json patter
      var streampattern = {
        "VtId": response.data['Vidinfo']['videotrack'][0]['Index'],
        "VidPreset": "",
        "AudPreset": "",
        "AudioT": [],
        "SubtitleT": []
      }

      localStorage.removeItem('filename');
      localStorage.setItem('video', JSON.stringify(response.data));
      localStorage.setItem('streampattern', JSON.stringify(streampattern));
    
    })
    .catch(function (error) {
      console.log(error)
      // handle error
    });
}

function transcode(event) {
  var data = {
    "FileName": localStorage.getItem('filename'),
    "Streams": []
  }

  var video = JSON.parse(localStorage.getItem('video'));

  for (var j = 1; j <= formGroupCount; j++) {
    var strpat = JSON.parse(localStorage.getItem('streampattern'));
    // Video preset
    var e = document.getElementById(`video-presets-${j}`);
    if (e.options[e.selectedIndex].value != "nochange"){
      strpat['VidPreset'] = e.options[e.selectedIndex].value;
    }

    // Audio preset
    e = document.getElementById(`audio-presets-${j}`);
    if (e.options[e.selectedIndex].value != "nochange"){
      strpat['AudPreset'] = e.options[e.selectedIndex].value;
    }
    

    // Audio tracks
    e = document.getElementById(`audio-select-${j}`);
    var audioindex = e.options[e.selectedIndex].value;
    if (audioindex != "null" && audioindex != "keep") {
      var atindex = video.Vidinfo.audiotrack.findIndex(item => item.index == audioindex);
      var audio = {
        "AtId": parseInt(audioindex),
        "Lang": video.Vidinfo['audiotrack'][atindex]['language']
      };
      strpat['AudioT'].push(audio)

    } else if (audioindex == "keep") {
      var audio = {
        "AtId": -1,
        "Lang": ""
      };
      strpat['AudioT'].push(audio)
    }

    // Subtitle tracks
    e = document.getElementById(`subtitle-select-${j}`);
    var subindex = e.options[e.selectedIndex].value;
    if (subindex != "null" && subindex != "keep") {
      var stindex = video.Vidinfo.subtitle.findIndex(item => item.index == subindex);
      var sub = {
        "StId": parseInt(subindex),
        "Lang": video.Vidinfo['subtitle'][stindex]['Language']
      };
      strpat['SubtitleT'].push(sub)

    } else if (subindex == "keep") {
      var sub = {
        "StId": -1,
        "Lang": ""
      };
      strpat['SubtitleT'].push(sub)
    }

    data['Streams'].push(strpat)
  }

  // Send client data to server
  axios.post('/transcode', data)
    .then(function (response) {
      localStorage.removeItem('cldata');
      localStorage.removeItem('video');
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
