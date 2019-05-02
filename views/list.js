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
                tempbt.className = 'list-item-watch';
                tempbt.innerHTML = 'Watch';
                tempbt.setAttribute('onclick', `watch(${i})`);

                tempa.appendChild(tempbt);
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
                tempbt.className = 'list-item-watch';
                tempbt.innerHTML = 'Watch';
                tempbt.setAttribute('onclick', `watch(${i})`);

                tempa.appendChild(tempbt);
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

function watch(i) {
    var nginxUrl = 'http://localhost:88/dash/';
    var nginxVodFile = '/manifest.mpd';

    localStorage.setItem('mnfst', nginxUrl + document.getElementById(`file-name-${i}`).innerText + nginxVodFile);

    document.location = "/watch";
}