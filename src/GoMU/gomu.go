package main

import (
	"GoMU/engine"
	"GoMU/network"
	"code.google.com/p/go.net/websocket"
	"net/http"
)

func main() {
	eng := engine.Init("../data/world.json")

	http.Handle("/", http.FileServer(http.Dir("../client")))

	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		network.Connection(eng, ws)
	}))

	http.ListenAndServe(":8080", nil)
}
