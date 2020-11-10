var HOST = "http://localhost:8080"
var ROOMS = []      //該使用者的聊天室名單
var MEMBERS = []    //所有使用者名單
var ONLINE = []
var RECIPIENT = ''  //私聊接收者

//取得聊天室ＩＤ	
var url = new URL(location.href)
var CHATROOM = location.pathname.replace("/chat/", "")

//取得使用者ＩＤ
var USER = url.searchParams.get('user')

//取得私訊與否
var PRIVATION = url.searchParams.get('private')

//列出url已有資訊+房間操作功能
var roomTitle = new Vue({
    el: '#room-title',
    data: {
        title: CHATROOM,
        seen_leave: true,
        seen_del: true,
    },
    methods: {
        DeleteRoom: function(){
            swalDelRoom(this.title)
        },
        LeaveRoom: function(){
            leaveRoom(this.title)
        },
        NewRoom: function(){
            newRoom()
        }
    }
})

var myUserBlock = new Vue({
    el: '#myUserBlock',
    data: {
        username: USER
    }
})

//列出聊天室清單 (資料庫結果)
var roomList = new Vue({
    el: '#roomList',
    data: {
        rooms: ROOMS
    },
    methods: {
        goToRoom(room_id){
            let search = replaceQueryParam('private', 'false', location.search)
            window.location.href = HOST + "/chat/" + room_id + search
        },
        inRoom(room_id){
            return (room_id === CHATROOM) ? inRoomSymb : '' 
        }
    },
    mounted() {
        axios
            .get(HOST + "/roomlist" + location.search)
            .then(function(e){
                let list = e.data.rooms.split(',')
                for( let i=0; i < list.length ;i++){
                    let tmp = {'id': i, 'room_id':list[i], 'len': 0}
                    ROOMS.push(tmp)
                }
            })
    },
    
})

//列出所有使用者
var allUserList = new Vue({
    el: '#all-users',
    data: {
        members: MEMBERS,
        seen: false,
        search: ''
    },
    methods: {
        privateChat(toWho){
            let sort_users = [USER, toWho].sort();
                makePrivateRoom(sort_users)
        }
    },
    computed: {
        // 搜尋並返回結果
        filterd: function(){
            var s = this.search.toLowerCase();
            return(s.trim() !== '') ?
                this.members.filter(function(d){ return d.username.toLowerCase().indexOf(s) > -1; }) :
                this.members
        }
    },
    mounted() {
        axios
            .get(HOST + "/userlist")
            .then(function(e){
                let list = e.data.users.split(',')
                for( let i=0; i < list.length ;i++){
                    if(list[i] == USER)   //是自己的話不用列出
                        continue

                    let tmp = {'id': i, 'username':list[i]}
                    MEMBERS.push(tmp)
                }
            })
    },

})

//列出在線使用者
var onlineUserList = new Vue({
    el: '#online-users',
    data: {
        seen: true,
        o_members: [],
    },
    methods: {
        changeOnline: function(list){
            this.o_members = list
        }
    }
})

var switchAllOnline = new Vue({
    el: '#switch-all-online',
    data: {
        onlineColor: '#413636',
        allColor: '#827a7a',
    },
    methods: {
        // 切換按鈕使用狀態與區塊顯示判斷
        sOnline: function(){
            onlineUserList.$data.seen = true
            allUserList.$data.seen = false
            this.onlineColor = '#413636'
            this.allColor = '#827a7a'
        },
        sAll: function(){
            onlineUserList.$data.seen = false
            allUserList.$data.seen = true
            this.onlineColor = '#827a7a'
            this.allColor = '#413636'
        }
    }
})

/*ws*/
var conn;
var log = document.getElementById("log");
var inRoomSymb = `<i class="fas fa-fish" style="margin-right:0.5em;color:#00798F"></i>`;

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
    var item = document.createElement('div');
    var chat = JSON.parse(message);
    var chatTime = new Date(chat.time).toLocaleString('zh-TW');
    //判斷是否為系統訊息
    if (chat.sender == "SYS") {

        //系統hint 使用者名單
        if (chat.type == "H") {
            info = JSON.parse(chat.content)
            ONLINE = []                  //清空原來的在線人員列表

            if (CHATROOM != info.room_info) {
                alert("聊天室位置出錯!" + CHATROOM + info.room_info);
            }

            var members = info.user_info.split(",")
            for( let i=0; i < members.length ;i++){
                if(members[i] == USER)   //是自己的話不用列出
                    continue

                let tmp = {'username':members[i]}                
                ONLINE.push(tmp)
            }

            // 更改並列出目前在線名單            
            onlineUserList.changeOnline(ONLINE)

        }
        else if (chat.type == "WP") {
            if (!CHATROOM.includes(chat.content) && PRIVATION != true){
                showToastr(chat.content)
            }
            
        }
        //系統info
        else {
            info = JSON.parse(chat.content)
            rooms = JSON.parse(info.room_info)  //聊天室名單對應人數
            users = info.user_info.split(',')   //聊天室所有在線人員

            /*聊天室人數變更*/
            let roomlist_states = (Object.keys(rooms))

            for (ROOM of ROOMS) {
                if (roomlist_states.includes(ROOM.room_id)) {
                    ROOM.len = rooms[ROOM.room_id]
                }
            }

        }
        //系統訊息ex.ＸＸＸ離開聊天室
        item.innerHTML =  `<div class="system-text"><label>` + info.text + `</label></div>`
    }
    else {
        //判別內容是否包含鏈結
        var text = isUrl(chat.content)

        //來自其他用戶或使用者的廣播消息
        if (chat.type == "A") {

            //是私訊的話把全域廣播擋下來
            if (PRIVATION == "true"){ 
                return
            }

            item.innerHTML =  `<div class="chat-text">\
                <label class="sm-text"><span style="font-weight: 1000;">`+ chat.sender +`</span> 於 ` + chatTime + `廣播</lable><br>\
                <label class="bro-text">&nbsp;&nbsp;` + text + `</label>\
            </div>`
        }
        //一般的頻道消息
        else {
            item.innerHTML =  `<div class="chat-text">\
                <label class="sm-text"><span style="font-weight: 1000;">` + chat.sender + `</span> 於` + chatTime + `</lable><br>\
                <label class="md-text">&nbsp;&nbsp;` + text + `</label>\
            </div>`
        }
    }
    //打印訊息            
    appendLog(item);

}

var chatForm = new Vue({
    el: '#form',
    data: {
        msg : '',
        type : 'N'
    },
    methods: {
        sendMsg: function(){
            var content = this.msg
            if (!conn) {
                return false 
            }
            if (this.msg = ''){
                return false
            }
            if (PRIVATION == "true"){
                this.type = "P"
            }
            jstr = JSON.stringify({ sender: USER, roomId: CHATROOM, recipient: RECIPIENT, type: this.type, content: content, time: Date.now() });
            conn.send(jstr)

            return false
        }
    }
})


/*other function*/

//判斷是否為大廳和私聊
if (PRIVATION == "true" || CHATROOM == "main") {
    if(PRIVATION == "true"){
        roomTitle.$data.title = "私聊：" + CHATROOM

        let members = CHATROOM.split("-")
        if (members[0] == USER) {
            RECIPIENT = members[1]
        }
        else {
            RECIPIENT = members[0]
        }
    }
    roomTitle.$data.seen_leave = false
    roomTitle.$data.seen_del = false
}

//替換參數
function replaceQueryParam(param, newval, search) {
    var regex = new RegExp("([?;&])" + param + "[^&;]*[;&]?");
    var query = search.replace(regex, "$1").replace(/&$/, '');

    return (query.length > 2 ? query + "&" : "?") + (newval ? param + "=" + newval : '');
}

//判斷是否為超連結
function isUrl(v){
    var reg = /(http:\/\/|https:\/\/)((\w|=|\?|\.|\/|&|#|-)+)/g;
    v = v.replace(reg, `<a href='$1$2' target="_blank">$1$2</a>`).replace(/\n/g, "<br />");
    return v
}

//將聊天訊息放入聊天區塊
function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);

    if (doScroll) {
        log.scrollTop = log.scrollHeight - log.clientHeight;
    }
}

//私訊通知＿toastr通知設定
function showToastr(id){
    toastr.options = {
        "closeButton": false,
        "debug": false,
        "newestOnTop": false,
        "progressBar": false,
        "positionClass": "toast-bottom-right",
        "preventDuplicates": false,
        "onclick": null,
        "showDuration": "300",
        "hideDuration": "1000",
        "timeOut": "3000",
        "extendedTimeOut": "1000",
        "showEasing": "swing",
        "hideEasing": "linear",
        "showMethod": "fadeIn",
        "hideMethod": "fadeOut"
    }
    toastr["info"]("您有來自"+id+"的私訊", "通知");
}

/*房間操作 (ajax)*/

//刪除房間中介點
function swalDelRoom (room_id) {
    swal({
        title: "刪除該聊天室？",
        text: "該聊天室資料與聊天記錄會全部消失",
        buttons: true,
        dangerMode: true,
    }).then((willDelete) => {
        if (willDelete) {
            delRoom(room_id)
        }
    })

}

//刪除房間（資料庫）
function delRoom(id) {
    axios({
        method: 'get',
        baseURL: HOST,
        url: "/delete/" + id + "?user=" + USER,
    }).then((res)=>{
        swal("成功刪除", id + "聊天室含淚跟你說再見", "success")
        window.location = res.request.responseURL
    }).catch((err)=>{
        swal(err + "出錯了！刪除失敗！", id + "聊天室陰魂不散～", "error")
    })
}

//離開房間（刪除房間與自己的關聯＿資料庫）
function leaveRoom(id) {
    axios({
        method: 'get',
        baseURL: HOST,
        url: "/leave/" + id + "?user=" + USER,
    }).then((res)=>{
        swal("您已退出聊天室", id + "裡的朋友們會想念你的", "success")
        setTimeout(()=>{window.location.href = res.request.responseURL}, 1000)
    }).catch((err)=>{
        swal("出錯了！", id + "聊天室不想與你分開～", "error")
    })
}

//新增房間
function newRoom () {
    swal({
        title: "建立/前往 聊天室",
        text: "聊天室id:",
        content: "input",
        buttons: {
            cancel: true,
            confirm: true,
        },
    }).then(function (inputValue) {
        if (inputValue === null) return false;
        if (inputValue === "") {
            sweetAlert("哎呦……", "請輸入聊天室id", "error");
            return false
        }
        if (inputValue.length > 30) {
            sweetAlert("太…長……啦", "聊天室id為30字元內", "warning");
            return false
        }

        makeNormalRoom(inputValue)
    });
}

// 新建聊天室
function makeNormalRoom(roomName) {
    var xhr = new XMLHttpRequest();
    $.ajax({
        type: 'POST',
        url: location.protocol + "/normalroom",
        data: { "user": USER, "roomName": roomName },
        xhr: function () {
            return xhr
        },
        success: function () {
            window.location.href = xhr.responseURL
        }
    })
}

// 建立私聊連結
function makePrivateRoom(s) {
    var xhr = new XMLHttpRequest();
    $.ajax({
        type: 'POST',
        url: location.protocol + "/privateroom",
        data: { "user": USER, "roomName": s[0] + "-" + s[1] },
        xhr: function () {
            return xhr
        },
        success: function () {
            window.location.href = xhr.responseURL
        }
    })
}