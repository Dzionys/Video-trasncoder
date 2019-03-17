(function (d, axios) {
    "use strict";
    var inputFile = d.querySelector("#input-file");
    var uploadForm = document.getElementById("upload-form")
    var uploadFormLabel = document.getElementById("upload-form-label")

    inputFile.addEventListener("change", addFile);

    function addFile(e) {
        var file = e.target.files[0]
        if(!file){
            return
        }
        upload(file);
    }

    function upload(file) {
        var formData = new FormData()
        formData.append("file", file)
        post("/upload", formData)
            .then(onResponse)
            .catch(onResponse);        
    }

    function onResponse(response) {
        var className = (response.status !== 400) ? "success" : "error";
        uploadForm.className = "upload-form uploaded";
        uploadFormLabel.style.height = uploadForm.clientHeight + "px";
        uploadFormLabel.style.height = "0px";
        uploadFormLabel.className ="upload-form-label uploaded";
        // divNotification.innerHTML = response.data;
    }
})(document, axios)