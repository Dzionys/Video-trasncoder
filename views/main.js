'use strict';

function SSE() {
  var source = new EventSource('/sse/dashboard');
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

var inputFile = document.getElementById('input-file');
var transcodeSubmit = document.getElementById('transcode-submit');
var uploadForm = document.getElementById('upload-form');
var transcodeForm = document.getElementById('transcode');
var uploadFormLabel = document.getElementById('upload-form-label');
var uploadFormInput = document.getElementById('input-file');

var codec = document.getElementById('codec');
var resolution = document.getElementById('resolution');
var audioTracks = document.getElementById('audio-tracks');
var subtitleTracks = document.getElementById('subtitle-tracks');

function upload(event) {
  var file = event.target.files[0];
  if (!file) {
    return;
  }

  var formData = new FormData();
  formData.append('file', file);
  axios.post('/upload', formData)
    .then(function (response) {
      localStorage.setItem('video', response.data);

      uploadForm.className = 'upload-form uploaded';
      uploadFormLabel.className = 'upload-form-label uploaded';
      transcodeForm.className = 'transcode-form uploaded';
      uploadFormInput.disabled = true;
      resolution.innerText = `${response.data['videotrack'][0]['width']}x${response.data['videotrack'][0]['height']}`;
      codec.innerText = response.data['videotrack'][0]['codecName'];

      response.data.audiotrack.map(function (value) {
        audioTracks.innerHTML += `<option selected value="${value.index}">${value.index}</option>`
      })

      response.data.subtitle.map(function (value) {
        subtitleTracks.innerHTML += `<option selected value="${value.index}">${value.index}</option>`
      })

    })
    .catch(function (error) {
      // handle error
    });
}

function transcode(event) {
  var formData = new FormData(transcodeForm);
  formData.append('filename', localStorage.getItem('filename'))

  axios.post('/transcode', formData)
    .then(function (response) {
      transcodeForm.className = 'transcode-form';
    })
    .catch(function (error) {
      // handle error
    });

  event.preventDefault();
}

inputFile.addEventListener('change', upload);
transcodeForm.addEventListener('submit', transcode);
