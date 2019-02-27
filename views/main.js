function onLoaded(){
    var source = new EventSource("sse/dashboard")
    var logg = "";
    var currentmsg = "";

    source.onmessage = function (event){
        var dashboard = event.data;
        if (dashboard != currentmsg){
            console.log("OnMessage called:");
            console.dir(event);
            currentmsg = dashboard;
            logg += currentmsg + "<br/>";
            console.log(logg);
            document.getElementById("console").innerHTML = logg;
        }
    }
}