// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.URL.Query().Get("user") == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther) //在這裡加私訊驗證
	}
	if r.URL.Query().Get("private") == "true" {
		user := r.URL.Query().Get("user")
		pusers := strings.Split(r.URL.Query().Get("room"), "-")
		if len(pusers) != 2 {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !(user == pusers[0] || user == pusers[1]) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "public/home.html")
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprint(w, "Form value error")
		return
	}

	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		log.Println(t.Execute(w, nil))
	} else {
		// 在這邊放驗證
		user := strings.Trim(r.Form["username"][0], " ")
		http.Redirect(w, r, "/?user="+user+"&room=main&private=false", http.StatusSeeOther) //進入聊天室大廳
	}
}

func makePrivateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := r.FormValue("user")
		roomName := r.FormValue("roomName")
		http.Redirect(w, r, "/?user="+user+"&room="+roomName+"&private=ture", http.StatusFound) //進入聊天室

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

func main() {
	flag.Parse()
	hub := newHub()
	//rdb := GetRedisClient()
	go hub.run()
	http.HandleFunc("/login", login)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		data := hub.makeInfo()
		fmt.Fprint(w, string(data))
	})
	http.HandleFunc("/privateroom", makePrivateRoom)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
