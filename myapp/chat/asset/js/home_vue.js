var HOST = "http://localhost:8080"
var ROOMS = []      //該使用者的聊天室名單
var MEMBERS = []    //所有使用者名單

//取得聊天室ＩＤ	
var url = new URL(location.href)
var chatRoom = location.pathname.replace("/chat/", "")

//取得使用者ＩＤ
var user = url.searchParams.get('user')

//取得私訊與否
var privation = url.searchParams.get('private')

//列出url已有資訊+房間操作功能
var roomTitle = new Vue({
    el: '#room-title',
    data: {
        title: chatRoom
    },
    methods: {
        DeleteRoom(room_id){
            swalDelRoom(room_id)
        },
        LeaveRoom(room_id){
            leaveRoom(room_id)
        },
        NewRoom: function(){
            newRoom()
        }
    }
})

var myUserBlock = new Vue({
    el: '#myUserBlock',
    data: {
        username: user
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
        seen: false
    },
    mounted() {
        axios
            .get(HOST + "/userlist")
            .then(function(e){
                let list = e.data.users.split(',')
                for( let i=0; i < list.length ;i++){
                    let tmp = {'id': i, 'username':list[i]}
                    MEMBERS.push(tmp)
                }
            })
    },

})

var onlineUserList = new Vue({
    el: '#online-users',
    data: {
        seen: true
    }
})

/*other function*/

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
    var xhr = new XMLHttpRequest();
    $.ajax({
        type: 'GET',
        url: location.protocol + "/delete/" + id + "?user=" + user,
        xhr: function () {
            return xhr
        },
        success: function () {
            swal("成功刪除", id + "聊天室含淚跟你說再見", "success")
            window.location.href = xhr.responseURL
        },
        error: function () {
            swal("出錯了！刪除失敗！", id + "聊天室陰魂不散～", "error")
        }
    })
}

//離開房間（刪除房間與自己的關聯＿資料庫）
function leaveRoom(id) {
    var xhr = new XMLHttpRequest();
    $.ajax({
        type: 'GET',
        url: location.protocol + "/leave/" + id + "?user=" + user,
        xhr: function () {
            return xhr
        },
        success: function () {
            swal("您已退出聊天室", id + "裡的朋友們會想念你的", "success")
            setTimeout(()=>{window.location.href = xhr.responseURL}, 1000)
            
        },
        error: function () {
            swal("出錯了！", id + "聊天室不想與你分開～", "error")
        }
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

        makeNormalRoom(user, inputValue)
    });
}

// 新建聊天室
function makeNormalRoom(user, roomName) {
    var xhr = new XMLHttpRequest();
    $.ajax({
        type: 'POST',
        url: location.protocol + "/normalroom",
        data: { "user": user, "roomName": roomName },
        xhr: function () {
            return xhr
        },
        success: function () {
            window.location.href = xhr.responseURL
        }
    })
}