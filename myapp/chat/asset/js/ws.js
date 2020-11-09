window.onload = function () {
    //WS連線：接收廣播訊息
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws" + location.pathname + location.search);
        // conn.debug = true;
        // conn.timeoutInterval = 3600;
        conn.onclose = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };
        conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                HandleMessage (messages[i])
            }
        }
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
    console.log("test")
}