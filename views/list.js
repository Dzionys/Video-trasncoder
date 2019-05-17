window.onload = list();

function list() {
    var data;
  
    axios.get('/vd')
    .then(function (response) {
        data = response.data['VideoStream'];

        if (data == null) {
            var temph4 = document.createElement('h4');
                temph4.id = 'file-name';
                temph4.className = 'file-name';
                temph4.innerHTML = 'No videos';
                document.getElementById('list').appendChild(temph4);

                return;
        }

        data.map((d, i) => {
            if (d['Stream']) {
                var newListItem = document.createElement('div');
                newListItem.id = `list-item-${i}`;
                newListItem.className = 'list-item';
                var vinfo = document.createElement('div');
                vinfo.style.display = 'none';
                vinfo.id = `vd-info-${i}`;
                vinfo.className = 'vd-info';

                var tempspan = document.createElement('span');
                tempspan.id = `file-status-${i}`;
                tempspan.className = 'list-item-status';
                tempspan.innerHTML = d['State'];
                var temph4 = document.createElement('h4');
                temph4.id = `file-name-${i}`;
                temph4.className = 'file-name';
                temph4.innerHTML = d['StreamName'];
                var tempib = document.createElement('button');
                tempib.innerHTML = 'Info';
                tempib.type = 'button';
                tempib.className = 'list-item-button';
                tempib.setAttribute('onclick', `showvidinfo(${i})`)
                var tempa = document.createElement('a');
                tempa.id = `watch-button-${i}`;
                var tempbt = document.createElement('button');
                tempbt.type = 'button';
                tempbt.className = 'list-item-button';
                tempbt.innerHTML = 'Watch';
                tempbt.setAttribute('onclick', `watch(${i})`);
                var tempdbt = document.createElement('button');
                tempdbt.type = 'button';
                tempdbt.className = 'list-item-button';
                tempdbt.innerHTML = 'Delete';
                tempdbt.setAttribute('onclick', `delet(${i}, true)`);
                var tempebt = document.createElement('button');
                tempebt.type = 'button';
                tempebt.className = 'list-item-button';
                tempebt.innerHTML = 'Edit';
                tempebt.setAttribute('onclick', `update(${i}, true)`);

                tempa.appendChild(tempbt);
                tempa.appendChild(tempdbt);
                tempa.appendChild(tempebt);
                tempa.appendChild(tempib);
                newListItem.appendChild(tempspan);
                newListItem.appendChild(temph4);
                newListItem.appendChild(tempa);

                document.getElementById('list').appendChild(newListItem);
                document.getElementById('list').appendChild(vinfo);
            } else {
                var newListItem = document.createElement('div');
                newListItem.id = `list-item-${i}`;
                newListItem.className = 'list-item';
                var vinfo = document.createElement('div');
                vinfo.style.display = 'none';
                vinfo.id = `vd-info-${i}`;
                vinfo.className = 'vd-info';

                var tempspan = document.createElement('span');
                tempspan.id = `file-status-${i}`;
                tempspan.className = 'list-item-status';
                tempspan.innerHTML = d['State'];
                var temph4 = document.createElement('h4');
                temph4.id = `file-name-${i}`;
                temph4.className = 'file-name';
                temph4.innerHTML = d['Video'][0].FileName;
                var tempib = document.createElement('button');
                tempib.innerHTML = 'Info';
                tempib.type = 'button';
                tempib.className = 'list-item-button';
                tempib.setAttribute('onclick', `showvidinfo(${i})`)
                var tempa = document.createElement('a');
                tempa.id = `watch-button-${i}`;
                var tempbt = document.createElement('button');
                tempbt.type = 'button';
                tempbt.className = 'list-item-button';
                if (d['State'] != 'Transcoded') {
                    if (d['State'] != 'Transcoding') {
                        tempbt.innerHTML = "Transcode";
                        tempbt.setAttribute('onclick', `transcode(${i})`)
                    }
                } else {
                    tempbt.innerHTML = 'Watch';
                    tempbt.setAttribute('onclick', `watch(${i})`);
                }
                var tempdbt = document.createElement('button');
                tempdbt.type = 'button';
                tempdbt.className = 'list-item-button';
                tempdbt.innerHTML = 'Delete';
                tempdbt.setAttribute('onclick', `delet(${i}, false)`);
                var tempebt = document.createElement('button');
                tempebt.type = 'button';
                tempebt.className = 'list-item-button';
                tempebt.innerHTML = 'Edit';
                tempebt.setAttribute('onclick', `update(${i}, false)`);

                if (d['State'] != "Transcoding") {
                    tempa.appendChild(tempbt);
                }
                tempa.appendChild(tempdbt);
                tempa.appendChild(tempebt);
                tempa.appendChild(tempib);
                newListItem.appendChild(tempspan);
                newListItem.appendChild(temph4);
                newListItem.appendChild(tempa);

                document.getElementById('list').appendChild(newListItem);
                document.getElementById('list').appendChild(vinfo);
            }
        });

    })
    .catch(function (error) {
        console.log(error);
    })
}

var showing = false

function showvidinfo(i) {
    var vinfodiv = document.getElementById(`vd-info-${i}`);
    if (showing == true) {
        vinfodiv.style.display = 'none';
        vinfodiv.innerHTML = '';
        showing = false;
        return
    }
    showing = true

    var name = document.getElementById(`file-name-${i}`).innerHTML;
    var th4 = document.createElement('h4');
    th4.className = 'vd-info-text';

    axios.get('/vd')
        .then(function (response) {
            var info='';
            data = response.data['VideoStream'];
            data.map((str, j) => {
                if (str['Stream'] && str['StreamName'] == name) {
                    str['Video'].map((d, i) => {
                        info += `
                        ${i+1} Video Name: ${d['FileName']}
                        <br>
                        Video Codec: ${d['VtCodec']}, 
                        Audio Codec: ${d['AudioT'][0]['AtCodec']},
                        Frame Rate: ${d['FrameRate']},
                        Resolution: ${d['VtRes']}
                        <br><br>
                        `;

                        console.log("1");
                    });
                } else if (!str['Stream'] && str['Video'][0]['FileName'] == name) {
                    info = `
                    Video Name: ${str['Video'][0]['FileName']}
                    <br>
                    Video Codec: ${str['Video'][0]['VtCodec']}, 
                    Audio Codec: ${str['Video'][0]['AudioT'][0]['AtCodec']},
                    Frame Rate: ${str['Video'][0]['FrameRate']},
                    Resolution: ${str['Video'][0]['VtRes']}
                    <br><br>
                    `;
                    console.log("2");
                }
            });

            th4.innerHTML = info;
            vinfodiv.appendChild(th4);
            vinfodiv.style.display = 'block';
        })
        .catch(function (error) {
            console.log(error)
            // handle error
        });

}

function transcode(i) {
    var name = document.getElementById(`file-name-${i}`).innerHTML;
    postvideoupdate(3, name, '', false);
}

function delet(i, stream) {
    var name = document.getElementById(`file-name-${i}`).innerHTML;
    var response = postvideoupdate(1, name, '', stream);

    if (response != null) {
        var listitem = document.getElementById(`list-item-${i}`);
        listitem.parentNode.removeChild(listitem);
    }
}

var updating = false;

function update(i, stream) {
    if (updating == true) {
        updating = false;
        var updateForm = document.getElementById(`item-update-${i}`);
        updateForm.parentNode.removeChild(updateForm);
        return
    }
    updating = true;
    var listItem = document.getElementById(`list-item-${i}`);
    var updateForm = document.createElement('form');
    updateForm.id = `item-update-${i}`;
    var inputField = document.createElement('input');
    inputField.id = `update-input-${i}`;
    var submit = document.createElement('input');
    submit.type = 'submit';
    submit.value = 'Submit';
    submit.className = "list-item-button";
    submit.setAttribute('onclick', `setvalue(${i}, ${stream})`);

    updateForm.appendChild(inputField);
    updateForm.appendChild(submit);
    listItem.appendChild(updateForm);
}

function setvalue(i, stream) {
    var value = document.getElementById(`update-input-${i}`).value;
    var ovalue = document.getElementById(`file-name-${i}`);
    var updateForm = document.getElementById(`item-update-${i}`);
    updateForm.parentNode.removeChild(updateForm);

    var response = postvideoupdate(2, value, ovalue.innerHTML, stream);

    if (response != null) {
        ovalue.innerHTML = value;
    }
}

function postvideoupdate(i, value, ovalue, stream) {
    var data = {
        "Utype": i,
        "Data": value,
        "Odata": ovalue,
        "Stream": stream
    }

    axios.post('/videoupdate', data)
        .then(function (response) {

        })
        .catch(function (error) {
            console.log(error)
            return null;
            // handle error
        });

        return 'ok';
}

function watch(i) {
    var nginxUrl = 'http://localhost:88/dash/';
    var nginxVodFile = '/manifest.mpd';

    localStorage.setItem('mnfst', nginxUrl + document.getElementById(`file-name-${i}`).innerText + nginxVodFile);

    document.location = "/watch";
}