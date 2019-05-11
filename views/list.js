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
                var tempspan = document.createElement('span');
                tempspan.id = `file-status-${i}`;
                tempspan.className = 'list-item-status';
                tempspan.innerHTML = d['State'];
                var temph4 = document.createElement('h4');
                temph4.id = `file-name-${i}`;
                temph4.className = 'file-name';
                temph4.innerHTML = d['StreamName'];
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
                newListItem.appendChild(tempspan);
                newListItem.appendChild(temph4);
                newListItem.appendChild(tempa);

                document.getElementById('list').appendChild(newListItem);
            } else {
                var newListItem = document.createElement('div');
                newListItem.id = `list-item-${i}`;
                newListItem.className = 'list-item';
                var tempspan = document.createElement('span');
                tempspan.id = `file-status-${i}`;
                tempspan.className = 'list-item-status';
                tempspan.innerHTML = d['State'];
                var temph4 = document.createElement('h4');
                temph4.id = `file-name-${i}`;
                temph4.className = 'file-name';
                temph4.innerHTML = d['Video'][0].FileName;
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
                newListItem.appendChild(tempspan);
                newListItem.appendChild(temph4);
                newListItem.appendChild(tempa);

                document.getElementById('list').appendChild(newListItem);
            }
        });

    })
    .catch(function (error) {
        console.log(error);
    })
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

function update(i, stream) {
    var listItem = document.getElementById(`list-item-${i}`);
    var updateForm = document.createElement('form');
    updateForm.id = `item-update-${i}`;
    var inputField = document.createElement('input');
    inputField.id = `update-input-${i}`;
    var submit = document.createElement('input');
    submit.type = 'submit';
    submit.value = 'Submit';
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