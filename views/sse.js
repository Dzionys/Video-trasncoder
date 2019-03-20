function onLoaded(){
    //Connecting to SSE server
    var source = new EventSource('/sse/dashboard');
    var logg = '';
    var currentmsg = '';

    source.onmessage = function (event){
        if (!event.data.startsWith('<')) {
            logg += '<span class="user">user@transcoder</span>:<span class="home">~</span>$ video-transcode ' + event.data + '<br>';
            document.getElementById("filename").innerText = event.data
        } else if (event.data.indexOf("Error") > -1) {
            logg += '<span class="error">' + event.data + '</span><br>';
        } else if (/^[\s\S]*<br>.*?Progress:.*?<br>$/.test(logg) && event.data.includes('Progress:')) {
            logg = logg.replace(/^([\s\S]*<br>)(.*?Progress:.*?)(<br>)$/, `$1${event.data}$3`)
        } else {
            currentmsg = event.data;
            logg += currentmsg + '<br>';
        }
        
        document.getElementById('console').innerHTML = logg;
    }
}