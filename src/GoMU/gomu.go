package main

import (
	"code.google.com/p/go.net/websocket"
	"net/http"
	"GoMU/network"
	"GoMU/engine"
)

func main() {
	eng := engine.Init("../data/world.json")

	http.Handle("/", http.FileServer(http.Dir("../client")))
	
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		network.Connection(eng, ws)
	}))

	http.ListenAndServe(":8080", nil)
}
