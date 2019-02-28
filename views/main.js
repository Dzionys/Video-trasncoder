function onLoaded(){
    //Connecting to SSE server
    var source = new EventSource("sse/dashboard")
    var logg = "";
    var currentmsg = "";

    source.onmessage = function (event){
        var dashboard = event.data;
        //If message changed printing it to console
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