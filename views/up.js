(function (d, axios) {
    'use strict';
    var inputFile = d.querySelector('#input-file');
    var uploadForm = document.getElementById('upload-form')
    var uploadFormLabel = document.getElementById('upload-form-label')
    var uploadFormInput = document.getElementById('input-file')

    var codec = document.getElementById('codec');
    var resolution = document.getElementById('resolution');

    inputFile.addEventListener('change', addFile);

    function addFile(e) {
        var file = e.target.files[0]
        if(!file){
            return
        }
        upload(file);
    }

    function upload(file) {
        var formData = new FormData()
        formData.append('file', file)
        post('/upload', formData)
            .then(onResponse)
            .catch(onResponse);        
    }

    function onResponse(response) {
        var className = (response.status !== 400) ? 'success' : 'error';
        uploadForm.className = 'upload-form uploaded';
        uploadFormLabel.className ='upload-form-label uploaded';
        uploadFormInput.disabled = true

        localStorage.setItem('video', response.data)

        resolution.innerText = `${response.data['videotrack'][0]['width']}x${response.data['videotrack'][0]['height']}`
        codec.innerText = response.data['videotrack'][0]['codecName']
    }
})(document, axios)