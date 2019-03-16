function onLoaded(){
    //Connecting to SSE server
    var source = new EventSource("/sse/dashboard");
    var logg = "";
    var currentmsg = "";

    source.onmessage = function (event){
        if (logg === "") {
            logg = '<span class="user">user@transcoder</span>:<span class="home">~</span>$ video-transcode ' + event.data + '<br/>';
        } else if (event.data.indexOf("Error") > -1) {
            logg += '<span class="error">' + event.data + '</span><br/>';
        } else {
            currentmsg = event.data;
            logg += currentmsg + "<br/>";
        }

        console.log("OnMessage called:");
        console.dir(event);
        console.log(logg);
        document.getElementById("console").innerHTML = logg;
    }
}