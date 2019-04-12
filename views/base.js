function sse() {
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

var inputFile = document.getElementById('input-file');
var transcodeSubmit = document.getElementById('transcode-submit');
var uploadForm = document.getElementById('upload-form');
var transcodeForm = document.getElementById('transcode');
var uploadFormLabel = document.getElementById('upload-form-label');
var uploadFormInput = document.getElementById('input-file');
var toggle = document.getElementById('toggle');
var checkBox = document.getElementById('checkBox');

var codec = document.getElementById('codec');
var resolution = document.getElementById('resolution');
var framerate = document.getElementById('frame-rate');
var audioTracks = document.getElementById('audio-tracks');
var subtitleTracks = document.getElementById('subtitle-tracks');

function reload() {
  location.reload();
}

document.addEventListener("DOMContentLoaded", function (event) {
  var _selector = document.querySelector('input[name=checkbox]');
  _selector.addEventListener('change', function (event) {
    var data = {
      "tc": true
    }
    if (_selector.checked) {
      axios.post('/tctype', data)
        .then(function (response) {
        })
        .catch(function (error) {
          console.log(error)
          // handle error
        })
    } else {
      data.typechange = false;
      axios.post('/tctype', data)
        .then(function (response) {
        })
        .catch(function (error) {
          console.log(error)
          // handle error
        })
    }
  });
});