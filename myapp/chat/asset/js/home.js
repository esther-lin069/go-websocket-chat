
window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");
    var inRoomSymb = `<i class="fas fa-fish" style="margin-right:0.5em;color:#00798F"></i>`;
    
    

    //將聊天訊息放入聊天區塊
    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        //log.appendChild(item);
        $("#log").append(item)
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }


    //傳送輸入框訊息to ws
    document.getElementById("form").onsubmit = function () {
        var type = "N"
        var recipient = ""
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }

        if ($("#msg-type").val() == "broadcast") {
            type = "A"
        }
        if (PRIVATION == "true") {
            type = "P"
            let members = CHATROOM.split("-")
            if (members[0] == USER) {
                recipient = members[1]
            }
            else {
                recipient = members[0]
            }

        }
        jstr = JSON.stringify({ sender: USER, roomId: CHATROOM, recipient: recipient, type: type, content: msg.value, time: Date.now() });
        conn.send(jstr)
        msg.value = "";
        return false;
    };

    //WS連線：接收廣播訊息
    if (window["WebSocket"]) {
        conn = new ReconnectingWebSocket("ws://" + document.location.host + "/ws" + location.pathname + location.search);
        conn.debug = true;
        conn.timeoutInterval = 3600;
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

    //處理訊息
    function HandleMessage (message){
        chat = JSON.parse(message);
        chatTime = new Date(chat.time).toLocaleString('zh-TW');
        //判斷是否為系統訊息
        if (chat.sender == "SYS") {

            //系統hint 使用者名單
            if (chat.type == "H") {
                $("#online-member-list").empty()
                info = JSON.parse(chat.content)

                if (CHATROOM != info.room_info) {
                    alert("聊天室位置出錯!" + CHATROOM + info.room_info);
                } //聊天室名稱
                var members = info.user_info.split(",")
                members.forEach(element => {
                    if (element == USER) {  //是自己的話就不用列出
                        return
                    }
                    //在線使用者名單
                    var box = $(`<dt>
                            <div class="card">
                                <div class="box">
                                <span style="color:#00798F;margin-right:8px;">
                                    <i class="fa fa-user"></i>
                                </span>
                                <span class="box-text">${element}</span> 
                                </div>
                            </div>
                        </dt>`)

                    $("#online-member-list").append(box)


                });
            }
            else if (chat.type == "WP") {
                if (!CHATROOM.includes(chat.content) && PRIVATION != true){
                    showToastr(chat.content)
                }
                
            }
            //系統info
            else {
                //$("#rooms dl").empty()  //清空聊天室列表
                $("#main-room").html("大廳") //清空大廳人數
                info = JSON.parse(chat.content)
                rooms = JSON.parse(info.room_info)
                users = info.user_info.split(',')
                //console.log(users)      //聊天室所有在線人員
                //console.log(rooms)      //聊天室名單對應人數

                /*聊天室人數變更*/
                let roomlist_states = (Object.keys(rooms))

                $("#rooms dt").each(function () {
                    let room_name = $(this).children().text()
                    room_name = room_name.substr(0, room_name.length - 4).replace(inRoomSymb, "")
                    //console.log(room_name)

                    if (roomlist_states.includes(room_name)) {
                        $(this).children().children("span").text(rooms[room_name])
                    }
                })


            }
            var item = $(`<div class="system-text"><label>${info.text}</label></div>`)
        }
        else {
            var text = isUrl(chat.content)
            if (chat.type == "A") {
                if (PRIVATION == "true"){ //是私訊的話把全域廣播擋下來
                    return
                }
                var bro_content = text;
                var item = $(`<div class="chat-text">
                <label class="sm-text"><span style="font-weight: 1000;">${chat.sender}</span> 於 ${chatTime} 廣播</lable><br>
                <label class="bro-text">&nbsp;&nbsp;${bro_content}</label>
            </div>`)
            }
            //一般的頻道消息
            else {
                var item = $(`<div class="chat-text">
                    <label class="sm-text"><span style="font-weight: 1000;">${chat.sender}</span> 於 ${chatTime}</lable><br>
                    <label class="md-text">&nbsp;&nbsp;${text}</label>
                </div>`)
            }
        }
        //打印訊息            
        appendLog(item);
    
    }


};